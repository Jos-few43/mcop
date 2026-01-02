package discovery

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"mcop/src/types"
)

// DiscoveryService handles discovery of MCP servers
type DiscoveryService struct {
	timeout time.Duration
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService() *DiscoveryService {
	return &DiscoveryService{
		timeout: 5 * time.Second,
	}
}

// ServerInfo represents discovered server information
type ServerInfo struct {
	ID          string
	Name        string
	URL         string
	Status      string
	Description string
	Tools       []string
	ResponseTime time.Duration
}

// DiscoverLocalServers discovers MCP servers running locally
func (d *DiscoveryService) DiscoverLocalServers() ([]ServerInfo, error) {
	var servers []ServerInfo
	
	// Check for common ports used by MCP servers
	ports := []int{3000, 3001, 8000, 8080, 9000, 9001}
	
	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%d", port)
		if d.isMCPServer(url) {
			serverInfo := ServerInfo{
				ID:          fmt.Sprintf("local_%d", port),
				Name:        fmt.Sprintf("Local MCP Server (Port %d)", port),
				URL:         url,
				Status:      "running",
				Description: fmt.Sprintf("MCP server running on localhost:%d", port),
				ResponseTime: d.getResponseTime(url),
			}
			
			// Try to get tools from the server
			tools, err := d.getServerTools(url)
			if err == nil {
				serverInfo.Tools = tools
			}
			
			servers = append(servers, serverInfo)
		}
	}
	
	return servers, nil
}

// DiscoverNetworkServers discovers MCP servers on the local network
func (d *DiscoveryService) DiscoverNetworkServers() ([]ServerInfo, error) {
	var servers []ServerInfo
	
	// Get local IP addresses
	localIPs, err := d.getLocalIPs()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IPs: %w", err)
	}
	
	// Check each IP address for common MCP ports
	for _, ip := range localIPs {
		// Skip localhost
		if ip == "127.0.0.1" || strings.Contains(ip, "::1") {
			continue
		}
		
		ports := []int{3000, 3001, 8000, 8080, 9000, 9001}
		
		for _, port := range ports {
			url := fmt.Sprintf("http://%s:%d", ip, port)
			if d.isMCPServer(url) {
				serverInfo := ServerInfo{
					ID:          fmt.Sprintf("network_%s_%d", strings.ReplaceAll(ip, ".", "_"), port),
					Name:        fmt.Sprintf("Network MCP Server (%s:%d)", ip, port),
					URL:         url,
					Status:      "running",
					Description: fmt.Sprintf("MCP server running on %s:%d", ip, port),
					ResponseTime: d.getResponseTime(url),
				}
				
				// Try to get tools from the server
				tools, err := d.getServerTools(url)
				if err == nil {
					serverInfo.Tools = tools
				}
				
				servers = append(servers, serverInfo)
			}
		}
	}
	
	return servers, nil
}

// DiscoverFromConfig discovers servers based on configuration
func (d *DiscoveryService) DiscoverFromConfig(configuredServers []types.MCPServer) ([]ServerInfo, error) {
	var servers []ServerInfo

	for _, configuredServer := range configuredServers {
		// Check if the server is a stdio-based server
		if strings.HasPrefix(configuredServer.URL, "stdio://") {
			// For stdio servers, we can't really discover them in the network sense
			// but we can represent them as available
			serverInfo := ServerInfo{
				ID:          configuredServer.ID,
				Name:        configuredServer.Name,
				URL:         configuredServer.URL,
				Status:      configuredServer.Status,
				Description: configuredServer.Description,
			}

			servers = append(servers, serverInfo)
		} else if strings.HasPrefix(configuredServer.URL, "http://") || strings.HasPrefix(configuredServer.URL, "https://") {
			// Check if the HTTP-based server is reachable
			if d.isMCPServer(configuredServer.URL) {
				serverInfo := ServerInfo{
					ID:          configuredServer.ID,
					Name:        configuredServer.Name,
					URL:         configuredServer.URL,
					Status:      "running",
					Description: configuredServer.Description,
					ResponseTime: d.getResponseTime(configuredServer.URL),
				}

				// Try to get tools from the server
				tools, err := d.getServerTools(configuredServer.URL)
				if err == nil {
					serverInfo.Tools = tools
				}

				servers = append(servers, serverInfo)
			}
		}
	}

	return servers, nil
}

// isMCPServer checks if the given URL is an MCP server
func (d *DiscoveryService) isMCPServer(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: d.timeout,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}
	
	// Add common headers that MCP servers might expect
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MCOP-Discovery/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	// Check if the response indicates this is an MCP server
	// This could be based on specific headers, status codes, or response content
	// For now, we'll just check for success responses and common MCP indicators
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// getResponseTime measures the response time of a server
func (d *DiscoveryService) getResponseTime(url string) time.Duration {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	
	client := &http.Client{
		Timeout: d.timeout,
	}
	
	start := time.Now()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return d.timeout // Return timeout duration if request creation fails
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return d.timeout // Return timeout duration if request fails
	}
	defer resp.Body.Close()
	
	return time.Since(start)
}

// getServerTools tries to get the tools from an MCP server
func (d *DiscoveryService) getServerTools(url string) ([]string, error) {
	// This would typically make an API call to the server to list its tools
	// For now, return an empty slice
	// In a real implementation, you would call an endpoint like /tools or make an MCP list_tools call
	return []string{}, nil
}

// getLocalIPs gets all local IP addresses
func (d *DiscoveryService) getLocalIPs() ([]string, error) {
	var ips []string
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // Skip down or loopback interfaces
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			
			if ip == nil || ip.IsLoopback() {
				continue
			}
			
			// Convert to 4-byte representation if IPv4
			ip = ip.To4()
			if ip == nil {
				continue // Skip IPv6 for now
			}
			
			ips = append(ips, ip.String())
		}
	}
	
	// Also add the hostname
	hostname, err := os.Hostname()
	if err == nil {
		ips = append(ips, hostname)
	}
	
	// Add localhost as well
	ips = append(ips, "localhost", "127.0.0.1")
	
	return ips, nil
}

// DiscoverAll discovers all available MCP servers using various methods
func (d *DiscoveryService) DiscoverAll(configuredServers []types.MCPServer) ([]ServerInfo, error) {
	var allServers []ServerInfo

	// Discover local servers
	localServers, err := d.DiscoverLocalServers()
	if err != nil {
		// Log the error but continue with other discovery methods
		fmt.Printf("Warning: failed to discover local servers: %v\n", err)
	} else {
		allServers = append(allServers, localServers...)
	}

	// Discover network servers
	networkServers, err := d.DiscoverNetworkServers()
	if err != nil {
		// Log the error but continue with other discovery methods
		fmt.Printf("Warning: failed to discover network servers: %v\n", err)
	} else {
		allServers = append(allServers, networkServers...)
	}

	// Discover from config
	configServers, err := d.DiscoverFromConfig(configuredServers)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Warning: failed to discover from config: %v\n", err)
	} else {
		allServers = append(allServers, configServers...)
	}

	// Remove duplicates
	uniqueServers := d.removeDuplicates(allServers)

	return uniqueServers, nil
}

// removeDuplicates removes duplicate servers based on URL
func (d *DiscoveryService) removeDuplicates(servers []ServerInfo) []ServerInfo {
	seen := make(map[string]bool)
	var unique []ServerInfo
	
	for _, server := range servers {
		if !seen[server.URL] {
			seen[server.URL] = true
			unique = append(unique, server)
		}
	}
	
	return unique
}

// PrintDiscoveredServers prints the discovered servers in a formatted way
func (d *DiscoveryService) PrintDiscoveredServers(servers []ServerInfo) {
	if len(servers) == 0 {
		fmt.Println("No MCP servers found.")
		return
	}
	
	fmt.Println("Discovered MCP Servers:")
	fmt.Println("=======================")
	
	for i, server := range servers {
		fmt.Printf("%d. %s\n", i+1, server.Name)
		fmt.Printf("   URL: %s\n", server.URL)
		fmt.Printf("   Status: %s\n", server.Status)
		fmt.Printf("   Response Time: %v\n", server.ResponseTime)
		if server.Description != "" {
			fmt.Printf("   Description: %s\n", server.Description)
		}
		if len(server.Tools) > 0 {
			fmt.Printf("   Tools: %s\n", strings.Join(server.Tools, ", "))
		}
		fmt.Println()
	}
}