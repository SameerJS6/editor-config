package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Path      string
	Files     int64
	Dirs      int64
	Size      int64
	ReadOrder int
	ModTime   time.Time
}

type Config struct {
	Root        string
	DryRun      bool
	AutoYes     bool
	ScanOnly    bool
	MinSize     int64
	MaxSize     int64
	OlderThan   int
	Interactive bool
	ExportPath  string
}

// Format bytes to human-readable size
func formatSize(size int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func formatNumber(num int64) string {
	if num == 0 {
		return "0"
	}

	// Add commas to the number
	numStr := fmt.Sprintf("%d", num)

	// Add commas every 3 digits from the right
	var result []byte
	for i, digit := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(digit))
	}
	return string(result)
}

func formatCount(count int64) string {
	formatted := formatNumber(count)

	var wordRep string
	floatCount := float64(count)

	switch {
	case count >= 1000000000: // Billion
		wordRep = fmt.Sprintf("(%.1f Billion)", floatCount/1000000000)
	case count >= 1000000: // Million
		wordRep = fmt.Sprintf("(%.1f Million)", floatCount/1000000)
	case count >= 1000: // Thousand
		wordRep = fmt.Sprintf("(%.1f Thousand)", floatCount/1000)
	default:
		return formatted
	}

	return fmt.Sprintf("%s %s", formatted, wordRep)
}

func dirSize(root string) Result {
	var res Result
	res.Path = root

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			res.Dirs++
			// Get modification time for the root directory
			if path == root {
				if info, err := d.Info(); err == nil {
					res.ModTime = info.ModTime()
				}
			}
		} else {
			info, err := d.Info()
			if err == nil {
				res.Files++
				res.Size += info.Size()
			}
		}
		return nil
	})

	if err != nil {
		return res
	}
	return res
}

func findNodeModules(root string) ([]string, error) {
	var nodeModules []string
	
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "node_modules" {
			nodeModules = append(nodeModules, path)
			return filepath.SkipDir
		}
		return nil
	})
	
	return nodeModules, err
}

func scanNodeModules(nodeModules []string) []Result {
	numWorkers := runtime.NumCPU()
	jobs := make(chan struct{ path string; index int }, len(nodeModules))
	results := make(chan Result, len(nodeModules))
	var wg sync.WaitGroup
	var processed int32
	total := len(nodeModules)

	if total > 10 { // Only show progress for larger scans
		fmt.Printf("📏 Processing %d directories...\n", total)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				res := dirSize(job.path)
				res.ReadOrder = job.index + 1
				results <- res

				// Update progress
				if total > 10 {
					current := atomic.AddInt32(&processed, 1)
					if current%10 == 0 || current == int32(total) {
						fmt.Printf("⏳ Progress: %d/%d directories processed\r", current, total)
					}
				}
			}
		}()
	}

	for i, dir := range nodeModules {
		jobs <- struct{ path string; index int }{path: dir, index: i}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []Result
	for r := range results {
		allResults = append(allResults, r)
	}

	if total > 10 {
		fmt.Printf("✅ All %d directories processed\n", total)
	}

	return allResults
}

// filterResults applies size and age filters to the results
func filterResults(results []Result, config Config) []Result {
	var filtered []Result

	for _, r := range results {
		if config.MinSize > 0 && r.Size < config.MinSize {
			continue
		}
		if config.MaxSize > 0 && r.Size > config.MaxSize {
			continue
		}

		if config.OlderThan > 0 {
			daysOld := int(time.Since(r.ModTime).Hours() / 24)
			if daysOld < config.OlderThan {
				continue
			}
		}

		filtered = append(filtered, r)
	}

	return filtered
}

// exportResults saves results to a JSON file
func exportResults(results []Result, exportPath string) error {
	type ExportResult struct {
		Path      string `json:"path"`
		Size      int64  `json:"size"`
		Files     int64  `json:"files"`
		Dirs      int64  `json:"dirs"`
		ReadOrder int    `json:"read_order"`
		ModTime   string `json:"modified_time"`
	}

	var exportData []ExportResult
	for _, r := range results {
		exportData = append(exportData, ExportResult{
			Path:      r.Path,
			Size:      r.Size,
			Files:     r.Files,
			Dirs:      r.Dirs,
			ReadOrder: r.ReadOrder,
			ModTime:   r.ModTime.Format(time.RFC3339),
		})
	}

	file, err := os.Create(exportPath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	fmt.Printf("📄 Results exported to: %s\n", exportPath)
	return nil
}

func deleteNodeModules(path string, dryRun bool) error {
	if dryRun {
		fmt.Printf("🔍 [DRY RUN] Would delete: %s\n", path)
		return nil
	}
	
	fmt.Printf("🗑️  Deleting: %s\n", path)
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("❌ Error deleting %s: %v\n", path, err)
		return err
	}
	fmt.Printf("✅ Deleted: %s\n", path)
	return nil
}

func deleteAllNodeModules(results []Result, dryRun bool) (int, int) {
	numWorkers := runtime.NumCPU()
	jobs := make(chan Result, len(results))
	var wg sync.WaitGroup
	var mu sync.Mutex
	deleted := 0
	failed := 0

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range jobs {
				err := deleteNodeModules(r.Path, dryRun)
				mu.Lock()
				if err != nil {
					failed++
				} else {
					deleted++
				}
				mu.Unlock()
			}
		}()
	}

	for _, r := range results {
		jobs <- r
	}
	close(jobs)

	wg.Wait()
	return deleted, failed
}

func askConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func interactiveSelection(results []Result) []Result {
	if len(results) == 0 {
		return results
	}

	fmt.Println("\n🎯 INTERACTIVE MODE")
	fmt.Println("Choose which node_modules directories to delete:")
	fmt.Println("Enter numbers separated by commas (e.g., '1,3,5'), 'all' for all, or 'none' to cancel")

	// Show simplified list for selection
	for i, r := range results {
		fmt.Printf("%2d. %s (%s)\n", i+1, filepath.Base(filepath.Dir(r.Path)), formatSize(r.Size))
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nYour choice: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("❌ Error reading input")
		return nil
	}

	input = strings.ToLower(strings.TrimSpace(input))

	if input == "none" || input == "cancel" {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	if input == "all" {
		return results
	}

	parts := strings.Split(input, ",")
	var selected []Result

	for _, part := range parts {
		part = strings.TrimSpace(part)
		var num int
		if _, err := fmt.Sscanf(part, "%d", &num); err == nil && num > 0 && num <= len(results) {
			selected = append(selected, results[num-1])
		}
	}

	if len(selected) == 0 {
		fmt.Println("❌ No valid selections made")
		return nil
	}

	fmt.Printf("✅ Selected %d directories for deletion\n", len(selected))
	return selected
}

func printSummary(results []Result) {
	var totalSize int64
	var totalFiles int64
	var totalDirs int64

	// Sort results by size (largest first)
	sortedResults := make([]Result, len(results))
	copy(sortedResults, results)
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Size > sortedResults[j].Size
	})

	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("📦 Found node_modules directories (sorted by size, largest first):")
	fmt.Println(strings.Repeat("=", 140))

	// Print table header
	fmt.Printf("%-8s %-8s %-12s %-12s %-12s %-12s %s\n",
		"Order", "Read", "Size", "Files", "Dirs", "Age", "Path")
	fmt.Println(strings.Repeat("-", 140))

	// Print table rows
	for i, r := range sortedResults {
		age := "unknown"
		if !r.ModTime.IsZero() {
			days := int(time.Since(r.ModTime).Hours() / 24)
			if days == 0 {
				age = "today"
			} else if days == 1 {
				age = "1 day"
			} else if days < 30 {
				age = fmt.Sprintf("%d days", days)
			} else if days < 365 {
				months := days / 30
				age = fmt.Sprintf("%d mo", months)
			} else {
				years := days / 365
				age = fmt.Sprintf("%d yr", years)
			}
		}
		fmt.Printf("%-8d %-8d %-12s %-12s %-12s %-12s %s\n",
			i+1, r.ReadOrder, formatSize(r.Size), formatNumber(r.Files), formatNumber(r.Dirs), age, r.Path)
		totalSize += r.Size
		totalFiles += r.Files
		totalDirs += r.Dirs
	}

	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("🧮 TOTAL: %d node_modules directories\n", len(results))
	fmt.Printf("💾 Total Size: %s\n", formatSize(totalSize))
	fmt.Printf("📄 Total Files: %s\n", formatCount(totalFiles))
	fmt.Printf("📁 Total Directories: %s\n", formatCount(totalDirs))
	fmt.Println(strings.Repeat("=", 140))
}

func main() {
	var config Config
	
	flag.StringVar(&config.Root, "dir", ".", "Directory to scan (default: current directory)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be deleted without actually deleting")
	flag.BoolVar(&config.AutoYes, "y", false, "Automatically answer yes to prompts (use with caution!)")
	flag.BoolVar(&config.ScanOnly, "scan", false, "Only scan and show results, don't delete")
	flag.Int64Var(&config.MinSize, "min-size", 0, "Minimum size threshold (bytes) - only process directories larger than this")
	flag.Int64Var(&config.MaxSize, "max-size", 0, "Maximum size threshold (bytes) - only process directories smaller than this")
	flag.IntVar(&config.OlderThan, "older-than", 0, "Only process directories older than X days")
	flag.BoolVar(&config.Interactive, "interactive", false, "Interactive mode - let user choose which directories to delete")
	flag.StringVar(&config.ExportPath, "export", "", "Export results to JSON file")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Scan, identify, and delete node_modules directories.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --scan                                    # Just scan and show results\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dir C:\\projects                        # Scan a specific directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dry-run                                # Show what would be deleted\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --min-size 1000000000                    # Only show dirs > 1GB\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --older-than 30                          # Only dirs older than 30 days\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --interactive                            # Choose which dirs to delete\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --export results.json --scan             # Export results to JSON\n", os.Args[0])
	}
	
	flag.Parse()
	
	start := time.Now()
	
	absPath, err := filepath.Abs(config.Root)
	if err != nil {
		fmt.Printf("❌ Error resolving path: %v\n", err)
		os.Exit(1)
	}
	
	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Printf("❌ Error accessing directory: %v\n", err)
		os.Exit(1)
	}
	
	if !info.IsDir() {
		fmt.Printf("❌ Error: %s is not a directory\n", absPath)
		os.Exit(1)
	}
	
	config.Root = absPath
	
	fmt.Printf("🚀 Scanning for node_modules in: %s\n", config.Root)
	fmt.Println()
	
	fmt.Println("🔍 Searching for node_modules directories...")
	nodeModules, err := findNodeModules(config.Root)
	if err != nil {
		fmt.Printf("❌ Error scanning directory: %v\n", err)
		os.Exit(1)
	}
	
	if len(nodeModules) == 0 {
		fmt.Println("✨ No node_modules directories found!")
		return
	}
	
	fmt.Printf("📊 Found %d node_modules directories\n", len(nodeModules))
	fmt.Println("📏 Calculating sizes (this may take a moment)...")
	
	results := scanNodeModules(nodeModules)

	// Apply filters if specified
	if config.MinSize > 0 || config.MaxSize > 0 || config.OlderThan > 0 {
		results = filterResults(results, config)
		fmt.Printf("🎯 After filtering: %d node_modules directories\n", len(results))
	}

	printSummary(results)

	// Export results if requested
	if config.ExportPath != "" {
		if err := exportResults(results, config.ExportPath); err != nil {
			fmt.Printf("❌ Export failed: %v\n", err)
		}
	}

	scanDuration := time.Since(start)
	fmt.Printf("⏱️  Scan completed in %.2f seconds\n\n", scanDuration.Seconds())

	if config.ScanOnly {
		return
	}

	if config.Interactive {
		results = interactiveSelection(results)
		if results == nil {
			return // User cancelled
		}
	} else if !config.AutoYes && !config.DryRun {
		if !askConfirmation("\n⚠️  Do you want to DELETE all these node_modules directories?") {
			fmt.Println("❌ Operation cancelled by user")
			return
		}
	}
	
	if config.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No files will be deleted")
	}
	
	fmt.Println()
	
	deleteStart := time.Now()
	deleted, failed := deleteAllNodeModules(results, config.DryRun)
	deleteDuration := time.Since(deleteStart)
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	if config.DryRun {
		fmt.Printf("🔍 DRY RUN SUMMARY:\n")
		fmt.Printf("   Would delete: %d directories\n", deleted)
	} else {
		fmt.Printf("✅ DELETION SUMMARY:\n")
		fmt.Printf("   Successfully deleted: %d directories\n", deleted)
		if failed > 0 {
			fmt.Printf("   Failed: %d directories\n", failed)
		}
	}
	fmt.Printf("⏱️  Deletion time: %.2f seconds\n", deleteDuration.Seconds())
	fmt.Printf("⏱️  Total time: %.2f seconds\n", time.Since(start).Seconds())
	fmt.Println(strings.Repeat("=", 80))
}