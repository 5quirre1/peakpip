package main
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
	"github.com/spf13/cobra"
)
const (
	PyPIURL         = "https://pypi.org/pypi"
	PyPISimpleURL   = "https://pypi.org/simple"
	DefaultTimeout  = 30 * time.Second
	MaxConcurrency  = 10
)
type Package struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Summary      string            `json:"summary"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Homepage     string            `json:"home_page"`
	License      string            `json:"license"`
	Keywords     []string          `json:"keywords"`
	Classifiers  []string          `json:"classifiers"`
	Dependencies []string          `json:"requires_dist"`
	URLs         map[string]string `json:"urls"`
}
type PackageInfo struct {
	Info     Package                `json:"info"`
	Releases map[string][]Release   `json:"releases"`
	URLs     []Release              `json:"urls"`
}
type Release struct {
	Filename     string `json:"filename"`
	URL          string `json:"url"`
	PackageType  string `json:"packagetype"`
	Size         int64  `json:"size"`
	MD5Digest    string `json:"md5_digest"`
	SHA256Digest string `json:"digests.sha256"`
	UploadTime   string `json:"upload_time"`
	PythonVersion string `json:"python_version"`
}
type PeakPip struct {
	client      *http.Client
	pythonPath  string
	pipPath     string
	concurrent  int
	quiet       bool
	verbose     bool
	dryRun      bool
	userInstall bool
	target      string
}
func NewPeakPip() *PeakPip {
	return &PeakPip{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		concurrent: MaxConcurrency,
	}
}
func (p *PeakPip) Initialize() error {
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		pythonPath, err = exec.LookPath("python")
		if err != nil {
			return fmt.Errorf("python executable not found in PATH")
		}
	}
	p.pythonPath = pythonPath
	pipPath, err := exec.LookPath("pip3")
	if err != nil {
		pipPath, err = exec.LookPath("pip")
		if err != nil {
			return fmt.Errorf("pip executable not found in PATH")
		}
	}
	p.pipPath = pipPath
	return nil
}
func (p *PeakPip) GetPackageInfo(packageName string) (*PackageInfo, error) {
	url := fmt.Sprintf("%s/%s/json", PyPIURL, packageName)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("package not found: %s", packageName)
	}
	var packageInfo PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode package info: %v", err)
	}
	return &packageInfo, nil
}
func (p *PeakPip) SearchPackages(query string) ([]Package, error) {
	url := fmt.Sprintf("%s/%s/", PyPISimpleURL, query)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		packageInfo, err := p.GetPackageInfo(query)
		if err != nil {
			return nil, err
		}
		return []Package{packageInfo.Info}, nil
	}
	return []Package{}, nil
}
func (p *PeakPip) InstallPackage(packageSpec string) error {
	if p.dryRun {
		fmt.Printf("would install: %s\n", packageSpec)
		return nil
	}
	args := []string{"install"}
	if p.quiet {
		args = append(args, "--quiet")
	}
	if p.verbose {
		args = append(args, "--verbose")
	}
	if p.userInstall {
		args = append(args, "--user")
	}
	if p.target != "" {
		args = append(args, "--target", p.target)
	}
	args = append(args, packageSpec)
	cmd := exec.Command(p.pipPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) UninstallPackage(packageName string) error {
	if p.dryRun {
		fmt.Printf("would uninstall: %s\n", packageName)
		return nil
	}
	args := []string{"uninstall", "-y"}
	if p.quiet {
		args = append(args, "--quiet")
	}
	if p.verbose {
		args = append(args, "--verbose")
	}
	args = append(args, packageName)
	cmd := exec.Command(p.pipPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) ListPackages(outdated bool) error {
	args := []string{"list"}
	if outdated {
		args = append(args, "--outdated")
	}
	if p.quiet {
		args = append(args, "--quiet")
	}
	if p.verbose {
		args = append(args, "--verbose")
	}
	cmd := exec.Command(p.pipPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) ShowPackage(packageName string) error {
	packageInfo, err := p.GetPackageInfo(packageName)
	if err != nil {
		return err
	}
	fmt.Printf("name: %s\n", packageInfo.Info.Name)
	fmt.Printf("version: %s\n", packageInfo.Info.Version)
	fmt.Printf("summary: %s\n", packageInfo.Info.Summary)
	fmt.Printf("author: %s\n", packageInfo.Info.Author)
	fmt.Printf("homepage: %s\n", packageInfo.Info.Homepage)
	fmt.Printf("license: %s\n", packageInfo.Info.License)
	if len(packageInfo.Info.Dependencies) > 0 {
		fmt.Printf("dependencies:\n")
		for _, dep := range packageInfo.Info.Dependencies {
			fmt.Printf("  %s\n", dep)
		}
	}
	if len(packageInfo.Info.Classifiers) > 0 {
		fmt.Printf("classifiers:\n")
		for _, classifier := range packageInfo.Info.Classifiers {
			fmt.Printf("  %s\n", classifier)
		}
	}
	return nil
}
func (p *PeakPip) UpgradePackage(packageName string) error {
	if p.dryRun {
		fmt.Printf("would upgrade: %s\n", packageName)
		return nil
	}
	args := []string{"install", "--upgrade"}
	if p.quiet {
		args = append(args, "--quiet")
	}
	if p.verbose {
		args = append(args, "--verbose")
	}
	if p.userInstall {
		args = append(args, "--user")
	}
	args = append(args, packageName)
	cmd := exec.Command(p.pipPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) DownloadPackage(packageName, destDir string) error {
	if p.dryRun {
		fmt.Printf("would download: %s to %s\n", packageName, destDir)
		return nil
	}
	args := []string{"download"}
	if destDir != "" {
		args = append(args, "--dest", destDir)
	}
	if p.quiet {
		args = append(args, "--quiet")
	}
	if p.verbose {
		args = append(args, "--verbose")
	}
	args = append(args, packageName)
	cmd := exec.Command(p.pipPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) FreezePackages() error {
	cmd := exec.Command(p.pipPath, "freeze")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) InstallRequirements(requirementsFile string) error {
	if p.dryRun {
		fmt.Printf("would install requirements from: %s\n", requirementsFile)
		return nil
	}
	args := []string{"install", "-r", requirementsFile}
	if p.quiet {
		args = append(args, "--quiet")
	}
	if p.verbose {
		args = append(args, "--verbose")
	}
	if p.userInstall {
		args = append(args, "--user")
	}
	if p.target != "" {
		args = append(args, "--target", p.target)
	}
	cmd := exec.Command(p.pipPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func (p *PeakPip) CheckPackage(packageName string) error {
	cmd := exec.Command(p.pipPath, "show", packageName)
	return cmd.Run()
}
func main() {
	peakPip := NewPeakPip()
	if err := peakPip.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing PeakPip: %v\n", err)
		os.Exit(1)
	}
	var rootCmd = &cobra.Command{
		Use:   "peakpip",
		Short: "peakpip - peaker pip trust",
		Long:  `peakpip is a faster (?) version of pip (it's a wrapper)`,
	}
	rootCmd.PersistentFlags().BoolVarP(&peakPip.quiet, "quiet", "q", false, "give less output")
	rootCmd.PersistentFlags().BoolVarP(&peakPip.verbose, "verbose", "v", false, "give more output")
	rootCmd.PersistentFlags().BoolVar(&peakPip.dryRun, "dry-run", false, "don't actually install anything, just print what would be done")
	rootCmd.PersistentFlags().IntVar(&peakPip.concurrent, "concurrent", MaxConcurrency, "number of concurrent operations")
	var installCmd = &cobra.Command{
		Use:   "install [package...]",
		Short: "install packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pkg := range args {
				if err := peakPip.InstallPackage(pkg); err != nil {
					return fmt.Errorf("failed to install %s: %v", pkg, err)
				}
			}
			return nil
		},
	}
	installCmd.Flags().BoolVarP(&peakPip.userInstall, "user", "U", false, "install to user directory")
	installCmd.Flags().StringVarP(&peakPip.target, "target", "t", "", "install packages into target directory")
	installCmd.Flags().StringP("requirements", "r", "", "install from requirements file")
	installCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		reqFile, _ := cmd.Flags().GetString("requirements")
		if reqFile != "" {
			return peakPip.InstallRequirements(reqFile)
		}
		return nil
	}
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall [package...]",
		Short: "uninstall packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pkg := range args {
				if err := peakPip.UninstallPackage(pkg); err != nil {
					return fmt.Errorf("failed to uninstall %s: %v", pkg, err)
				}
			}
			return nil
		},
	}
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "list installed packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			outdated, _ := cmd.Flags().GetBool("outdated")
			return peakPip.ListPackages(outdated)
		},
	}
	listCmd.Flags().Bool("outdated", false, "List outdated packages")
	var showCmd = &cobra.Command{
		Use:   "show [package...]",
		Short: "show information about packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pkg := range args {
				if err := peakPip.ShowPackage(pkg); err != nil {
					return fmt.Errorf("failed to show %s: %v", pkg, err)
				}
				if len(args) > 1 {
					fmt.Println("---")
				}
			}
			return nil
		},
	}
	var searchCmd = &cobra.Command{
		Use:   "search [query]",
		Short: "search for packages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packages, err := peakPip.SearchPackages(args[0])
			if err != nil {
				return err
			}
			for _, pkg := range packages {
				fmt.Printf("%s (%s) - %s\n", pkg.Name, pkg.Version, pkg.Summary)
			}
			return nil
		},
	}
	var upgradeCmd = &cobra.Command{
		Use:   "upgrade [package...]",
		Short: "upgrade packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pkg := range args {
				if err := peakPip.UpgradePackage(pkg); err != nil {
					return fmt.Errorf("failed to upgrade %s: %v", pkg, err)
				}
			}
			return nil
		},
	}
	var downloadCmd = &cobra.Command{
		Use:   "download [package...]",
		Short: "download packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			destDir, _ := cmd.Flags().GetString("dest")
			for _, pkg := range args {
				if err := peakPip.DownloadPackage(pkg, destDir); err != nil {
					return fmt.Errorf("failed to download %s: %v", pkg, err)
				}
			}
			return nil
		},
	}
	downloadCmd.Flags().StringP("dest", "d", "", "download directory")
	var freezeCmd = &cobra.Command{
		Use:   "freeze",
		Short: "output installed packages in requirements format",
		RunE: func(cmd *cobra.Command, args []string) error {
			return peakPip.FreezePackages()
		},
	}
	var checkCmd = &cobra.Command{
		Use:   "check [package...]",
		Short: "check if packages are installed",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pkg := range args {
				if err := peakPip.CheckPackage(pkg); err != nil {
					fmt.Printf("%s: not installed\n", pkg)
				} else {
					fmt.Printf("%s: installed\n", pkg)
				}
			}
			return nil
		},
	}
	rootCmd.AddCommand(installCmd, uninstallCmd, listCmd, showCmd, searchCmd, upgradeCmd, downloadCmd, freezeCmd, checkCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
