package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DependencyScanner handles dependency vulnerability scanning
type DependencyScanner struct {
	config *DependencyScanConfig
}

// DependencyScanConfig holds configuration for dependency scanning
type DependencyScanConfig struct {
	EnableGoModScan       bool
	EnableNpmScan         bool
	ScanInterval          time.Duration
	VulnerabilityDBURL    string
	AlertOnHighSeverity   bool
	AlertOnMediumSeverity bool
}

// Vulnerability represents a dependency vulnerability
type Vulnerability struct {
	ID          string    `json:"id"`
	Package     string    `json:"package"`
	Version     string    `json:"version"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CVE         string    `json:"cve,omitempty"`
	CVSS        float64   `json:"cvss,omitempty"`
	FixedIn     string    `json:"fixed_in,omitempty"`
	References  []string  `json:"references,omitempty"`
	DetectedAt  time.Time `json:"detected_at"`
}

// ScanResult represents the result of a dependency scan
type ScanResult struct {
	ScanID          string          `json:"scan_id"`
	ScanType        string          `json:"scan_type"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	TotalPackages   int             `json:"total_packages"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	HighSeverity    int             `json:"high_severity"`
	MediumSeverity  int             `json:"medium_severity"`
	LowSeverity     int             `json:"low_severity"`
	Status          string          `json:"status"`
	ErrorMessage    string          `json:"error_message,omitempty"`
}

// NewDependencyScanner creates a new dependency scanner
func NewDependencyScanner(config *DependencyScanConfig) *DependencyScanner {
	if config == nil {
		config = DefaultDependencyScanConfig()
	}
	return &DependencyScanner{
		config: config,
	}
}

// DefaultDependencyScanConfig returns default dependency scan configuration
func DefaultDependencyScanConfig() *DependencyScanConfig {
	return &DependencyScanConfig{
		EnableGoModScan:       true,
		EnableNpmScan:         true,
		ScanInterval:          24 * time.Hour, // Daily scans
		VulnerabilityDBURL:    "https://osv.dev/v1/query",
		AlertOnHighSeverity:   true,
		AlertOnMediumSeverity: false,
	}
}

// ScanDependencies performs a comprehensive dependency vulnerability scan
func (ds *DependencyScanner) ScanDependencies(ctx context.Context, projectPath string) (*ScanResult, error) {
	result := &ScanResult{
		ScanID:    fmt.Sprintf("scan_%d", time.Now().Unix()),
		StartTime: time.Now(),
		Status:    "running",
	}

	var allVulnerabilities []Vulnerability

	// Scan Go dependencies
	if ds.config.EnableGoModScan {
		goVulns, err := ds.scanGoModules(ctx, projectPath)
		if err != nil {
			result.Status = "error"
			result.ErrorMessage = fmt.Sprintf("Go scan failed: %v", err)
			return result, err
		}
		allVulnerabilities = append(allVulnerabilities, goVulns...)
		result.ScanType = "go"
	}

	// Scan NPM dependencies (if package.json exists)
	if ds.config.EnableNpmScan {
		npmVulns, err := ds.scanNpmPackages(ctx, projectPath)
		if err != nil {
			// Don't fail the entire scan if NPM scan fails (might not have Node.js project)
			fmt.Printf("NPM scan warning: %v\n", err)
		} else {
			allVulnerabilities = append(allVulnerabilities, npmVulns...)
			if result.ScanType == "" {
				result.ScanType = "npm"
			} else {
				result.ScanType = "mixed"
			}
		}
	}

	// Process results
	result.Vulnerabilities = allVulnerabilities
	result.TotalPackages = ds.countUniquePackages(allVulnerabilities)

	// Count by severity
	for _, vuln := range allVulnerabilities {
		switch strings.ToUpper(vuln.Severity) {
		case "HIGH", "CRITICAL":
			result.HighSeverity++
		case "MEDIUM", "MODERATE":
			result.MediumSeverity++
		case "LOW":
			result.LowSeverity++
		}
	}

	result.EndTime = time.Now()
	result.Status = "completed"

	return result, nil
}

// scanGoModules scans Go module dependencies for vulnerabilities
func (ds *DependencyScanner) scanGoModules(ctx context.Context, projectPath string) ([]Vulnerability, error) {
	// Check if go.mod exists
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("go.mod not found in %s", projectPath)
	}

	// Use govulncheck if available, otherwise fall back to go list
	vulnerabilities := []Vulnerability{}

	// Try govulncheck first
	if ds.hasGovulncheck() {
		vulns, err := ds.runGovulncheck(ctx, projectPath)
		if err == nil {
			return vulns, nil
		}
		fmt.Printf("govulncheck failed, falling back to manual scan: %v\n", err)
	}

	// Fall back to manual dependency listing and OSV API
	deps, err := ds.getGoDependencies(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Go dependencies: %w", err)
	}

	for _, dep := range deps {
		vulns, err := ds.queryOSVDatabase(ctx, dep.Name, dep.Version)
		if err != nil {
			fmt.Printf("Failed to query OSV for %s: %v\n", dep.Name, err)
			continue
		}
		vulnerabilities = append(vulnerabilities, vulns...)
	}

	return vulnerabilities, nil
}

// scanNpmPackages scans NPM package dependencies for vulnerabilities
func (ds *DependencyScanner) scanNpmPackages(ctx context.Context, projectPath string) ([]Vulnerability, error) {
	// Check if package.json exists
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package.json not found in %s", projectPath)
	}

	// Use npm audit if available
	vulnerabilities := []Vulnerability{}

	if ds.hasNpmAudit() {
		vulns, err := ds.runNpmAudit(ctx, projectPath)
		if err == nil {
			return vulns, nil
		}
		fmt.Printf("npm audit failed, falling back to manual scan: %v\n", err)
	}

	return vulnerabilities, nil
}

// hasGovulncheck checks if govulncheck is available
func (ds *DependencyScanner) hasGovulncheck() bool {
	_, err := exec.LookPath("govulncheck")
	return err == nil
}

// hasNpmAudit checks if npm audit is available
func (ds *DependencyScanner) hasNpmAudit() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// runGovulncheck runs govulncheck and parses results
func (ds *DependencyScanner) runGovulncheck(ctx context.Context, projectPath string) ([]Vulnerability, error) {
	cmd := exec.CommandContext(ctx, "govulncheck", "-json", "./...")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("govulncheck failed: %w", err)
	}

	// Parse govulncheck JSON output
	vulnerabilities := []Vulnerability{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		// Parse vulnerability from govulncheck output
		if result["type"] == "vuln" {
			vuln := ds.parseGovulncheckVuln(result)
			if vuln != nil {
				vulnerabilities = append(vulnerabilities, *vuln)
			}
		}
	}

	return vulnerabilities, nil
}

// runNpmAudit runs npm audit and parses results
func (ds *DependencyScanner) runNpmAudit(ctx context.Context, projectPath string) ([]Vulnerability, error) {
	cmd := exec.CommandContext(ctx, "npm", "audit", "--json")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		// npm audit returns non-zero exit code when vulnerabilities are found
		// So we need to check if we got output despite the error
		if len(output) == 0 {
			return nil, fmt.Errorf("npm audit failed: %w", err)
		}
	}

	// Parse npm audit JSON output
	var auditResult map[string]interface{}
	if err := json.Unmarshal(output, &auditResult); err != nil {
		return nil, fmt.Errorf("failed to parse npm audit output: %w", err)
	}

	vulnerabilities := []Vulnerability{}

	// Parse vulnerabilities from npm audit output
	if advisories, ok := auditResult["advisories"].(map[string]interface{}); ok {
		for _, advisory := range advisories {
			vuln := ds.parseNpmAuditVuln(advisory)
			if vuln != nil {
				vulnerabilities = append(vulnerabilities, *vuln)
			}
		}
	}

	return vulnerabilities, nil
}

// parseGovulncheckVuln parses a vulnerability from govulncheck output
func (ds *DependencyScanner) parseGovulncheckVuln(result map[string]interface{}) *Vulnerability {
	vuln := &Vulnerability{
		DetectedAt: time.Now(),
	}

	if id, ok := result["id"].(string); ok {
		vuln.ID = id
	}

	if pkg, ok := result["package"].(string); ok {
		vuln.Package = pkg
	}

	if summary, ok := result["summary"].(string); ok {
		vuln.Title = summary
	}

	if details, ok := result["details"].(string); ok {
		vuln.Description = details
	}

	// Determine severity (govulncheck doesn't provide severity, so we estimate)
	vuln.Severity = "MEDIUM" // Default severity

	return vuln
}

// parseNpmAuditVuln parses a vulnerability from npm audit output
func (ds *DependencyScanner) parseNpmAuditVuln(advisory interface{}) *Vulnerability {
	adv, ok := advisory.(map[string]interface{})
	if !ok {
		return nil
	}

	vuln := &Vulnerability{
		DetectedAt: time.Now(),
	}

	if id, ok := adv["id"].(float64); ok {
		vuln.ID = fmt.Sprintf("npm-%d", int(id))
	}

	if title, ok := adv["title"].(string); ok {
		vuln.Title = title
	}

	if overview, ok := adv["overview"].(string); ok {
		vuln.Description = overview
	}

	if severity, ok := adv["severity"].(string); ok {
		vuln.Severity = strings.ToUpper(severity)
	}

	if moduleName, ok := adv["module_name"].(string); ok {
		vuln.Package = moduleName
	}

	if cves, ok := adv["cves"].([]interface{}); ok && len(cves) > 0 {
		if cve, ok := cves[0].(string); ok {
			vuln.CVE = cve
		}
	}

	return vuln
}

// Dependency represents a package dependency
type Dependency struct {
	Name    string
	Version string
}

// getGoDependencies gets list of Go dependencies
func (ds *DependencyScanner) getGoDependencies(ctx context.Context, projectPath string) ([]Dependency, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-json", "all")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Go modules: %w", err)
	}

	dependencies := []Dependency{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var module map[string]interface{}
		if err := json.Unmarshal([]byte(line), &module); err != nil {
			continue
		}

		if path, ok := module["Path"].(string); ok {
			version := ""
			if v, ok := module["Version"].(string); ok {
				version = v
			}
			dependencies = append(dependencies, Dependency{
				Name:    path,
				Version: version,
			})
		}
	}

	return dependencies, nil
}

// queryOSVDatabase queries the OSV vulnerability database
func (ds *DependencyScanner) queryOSVDatabase(ctx context.Context, packageName, version string) ([]Vulnerability, error) {
	query := map[string]interface{}{
		"package": map[string]string{
			"name": packageName,
		},
	}

	if version != "" {
		query["version"] = version
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ds.config.VulnerabilityDBURL, strings.NewReader(string(queryJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query OSV: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var osvResponse map[string]interface{}
	if err := json.Unmarshal(body, &osvResponse); err != nil {
		return nil, fmt.Errorf("failed to parse OSV response: %w", err)
	}

	vulnerabilities := []Vulnerability{}

	if vulns, ok := osvResponse["vulns"].([]interface{}); ok {
		for _, v := range vulns {
			vuln := ds.parseOSVVuln(v, packageName, version)
			if vuln != nil {
				vulnerabilities = append(vulnerabilities, *vuln)
			}
		}
	}

	return vulnerabilities, nil
}

// parseOSVVuln parses a vulnerability from OSV API response
func (ds *DependencyScanner) parseOSVVuln(v interface{}, packageName, version string) *Vulnerability {
	vuln_data, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	vuln := &Vulnerability{
		Package:    packageName,
		Version:    version,
		DetectedAt: time.Now(),
		Severity:   "MEDIUM", // Default severity
	}

	if id, ok := vuln_data["id"].(string); ok {
		vuln.ID = id
	}

	if summary, ok := vuln_data["summary"].(string); ok {
		vuln.Title = summary
	}

	if details, ok := vuln_data["details"].(string); ok {
		vuln.Description = details
	}

	// Parse aliases for CVE
	if aliases, ok := vuln_data["aliases"].([]interface{}); ok {
		for _, alias := range aliases {
			if aliasStr, ok := alias.(string); ok && strings.HasPrefix(aliasStr, "CVE-") {
				vuln.CVE = aliasStr
				break
			}
		}
	}

	return vuln
}

// countUniquePackages counts unique packages in vulnerabilities
func (ds *DependencyScanner) countUniquePackages(vulnerabilities []Vulnerability) int {
	packages := make(map[string]bool)
	for _, vuln := range vulnerabilities {
		packages[vuln.Package] = true
	}
	return len(packages)
}
