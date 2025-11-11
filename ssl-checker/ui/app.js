// Parse target input to extract domain/IP and port
function parseTarget(target) {
    const trimmed = target.trim();
    if (!trimmed) {
        return null;
    }
    
    // First, check for bracketed IPv6 address with port: [IPv6]:port
    const ipv6BracketMatch = trimmed.match(/^\[([^\]]+)\]:(\d+)$/);
    if (ipv6BracketMatch) {
        const host = ipv6BracketMatch[1];
        const port = parseInt(ipv6BracketMatch[2], 10);
        
        // Validate port range
        if (port < 1 || port > 65535) {
            throw new Error('Port must be between 1 and 65535');
        }
        
        return {
            host: host,
            port: port
        };
    }
    
    // Check for bracketed IPv6 address without port: [IPv6]
    const ipv6BracketOnlyMatch = trimmed.match(/^\[([^\]]+)\]$/);
    if (ipv6BracketOnlyMatch) {
        return {
            host: ipv6BracketOnlyMatch[1],
            port: 443
        };
    }
    
    // Check for non-bracketed input with port: host:port (IPv4 or hostname)
    // Use a non-greedy pattern that matches host without colons, then :port
    const hostPortMatch = trimmed.match(/^([^:]+):(\d+)$/);
    
    if (hostPortMatch) {
        const host = hostPortMatch[1];
        const port = parseInt(hostPortMatch[2], 10);
        
        // Validate port range
        if (port < 1 || port > 65535) {
            throw new Error('Port must be between 1 and 65535');
        }
        
        return {
            host: host,
            port: port
        };
    }
    
    // No port specified, use default (could be IPv4, IPv6, or hostname)
    return {
        host: trimmed,
        port: 443
    };
}

// Determine if a host is an IP address (IPv4 or IPv6)
function isIPAddress(host) {
    // Strict IPv4 pattern - validates each octet is 0-255
    const ipv4Pattern = /^(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])$/;
    
    // IPv6 pattern - supports full, compressed, and IPv4-mapped forms
    // Matches standard IPv6, compressed (::), and mixed IPv4-in-IPv6 formats
    const ipv6Pattern = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;
    
    return ipv4Pattern.test(host) || ipv6Pattern.test(host);
}

// Single Check Form Handler
document.getElementById('singleCheckForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const form = e.target;
    const button = form.querySelector('button[type="submit"]');
    const btnText = button.querySelector('.btn-text');
    const spinner = button.querySelector('.spinner');
    
    const targetInput = form.target.value.trim();
    
    // Validation
    if (!targetInput) {
        alert('Please provide a domain name or IP address');
        return;
    }
    
    let parsed;
    try {
        parsed = parseTarget(targetInput);
    } catch (error) {
        alert(error.message);
        return;
    }
    
    if (!parsed) {
        alert('Please provide a valid domain name or IP address');
        return;
    }
    
    // Show loading state
    button.disabled = true;
    btnText.style.display = 'none';
    spinner.style.display = 'inline-block';
    
    try {
        const params = new URLSearchParams({ port: parsed.port });
        
        // Determine if host is IP or domain
        if (isIPAddress(parsed.host)) {
            params.append('ip', parsed.host);
        } else {
            params.append('domain', parsed.host);
        }
        
        const response = await fetch(`/api/check?${params.toString()}`);
        const data = await response.json();
        
        displayResults([data]);
    } catch (error) {
        console.error('Error:', error);
        displayError('Failed to check SSL certificate. Please try again.');
    } finally {
        // Reset button state
        button.disabled = false;
        btnText.style.display = 'inline';
        spinner.style.display = 'none';
    }
});

// Display Results
function displayResults(results) {
    const resultsDiv = document.getElementById('results');
    const resultsContent = document.getElementById('resultsContent');
    
    resultsContent.innerHTML = '';
    
    results.forEach((result, index) => {
        const resultItem = createResultItem(result, index);
        resultsContent.appendChild(resultItem);
    });
    
    resultsDiv.style.display = 'block';
    resultsDiv.scrollIntoView({ behavior: 'smooth' });
}

// Create Result Item
function createResultItem(result, index) {
    const div = document.createElement('div');
    div.className = 'result-item';
    
    if (result.status === 'error') {
        div.innerHTML = `
            <div class="result-header">
                <div class="result-title">Result ${index + 1}</div>
                <span class="status-badge status-error">Error</span>
            </div>
            <div class="error-message">
                ${escapeHtml(result.error || 'An error occurred')}
            </div>
        `;
        return div;
    }
    
    const data = result.data;
    const ssl = data.ssl || {};
    const ipInfo = data.ip_info || {};
    
    // Build HTML
    let html = `
        <div class="result-header">
            <div class="result-title">${escapeHtml(data.domain || data.ip)}</div>
            <span class="status-badge status-${data.sslStatus}">${data.sslStatus}</span>
        </div>
    `;
    
    // SSL Information
    if (ssl) {
        html += `
            <div class="result-section">
                <h3>🔒 SSL Certificate</h3>
                <div class="info-grid">`;
        
        // Display all subject fields
        if (ssl.subject) {
            // Create array of all subject properties
            const subjectEntries = Object.entries(ssl.subject);
            if (subjectEntries.length > 0) {
                subjectEntries.forEach(([key, value]) => {
                    // Format the label nicely
                    let label = key;
                    if (key === 'commonName') label = 'Subject CN (Common Name)';
                    else if (key === 'organizationName') label = 'Subject Organization';
                    else if (key === 'organizationalUnitName') label = 'Subject Organizational Unit';
                    else if (key === 'countryName') label = 'Subject Country';
                    else if (key === 'stateOrProvinceName') label = 'Subject State/Province';
                    else if (key === 'localityName') label = 'Subject Locality';
                    else label = `Subject ${key}`;
                    
                    html += `
                    <div class="info-item">
                        <div class="info-label">${escapeHtml(label)}</div>
                        <div class="info-value">${escapeHtml(value || 'N/A')}</div>
                    </div>`;
                });
            } else {
                html += `
                    <div class="info-item">
                        <div class="info-label">Subject CN</div>
                        <div class="info-value">N/A</div>
                    </div>`;
            }
        } else {
            html += `
                    <div class="info-item">
                        <div class="info-label">Subject CN</div>
                        <div class="info-value">N/A</div>
                    </div>`;
        }
        
        html += `
                    <div class="info-item">
                        <div class="info-label">Issuer</div>
                        <div class="info-value">${escapeHtml(ssl.issuer?.commonName || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Valid From</div>
                        <div class="info-value">${escapeHtml(ssl.notBefore || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Valid Until</div>
                        <div class="info-value">${escapeHtml(ssl.notAfter || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Days Until Expiration</div>
                        <div class="info-value">${ssl.daysUntilExpiration !== null ? ssl.daysUntilExpiration : 'N/A'}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">TLS Version</div>
                        <div class="info-value">${escapeHtml(ssl.tlsVersion || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Cipher Suite</div>
                        <div class="info-value">${escapeHtml(ssl.cipherSuite || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Signature Algorithm</div>
                        <div class="info-value">${escapeHtml(ssl.signatureAlgorithm || 'N/A')}</div>
                    </div>
                </div>
            </div>
        `;
        
        // Alerts
        if (ssl.alerts && ssl.alerts.length > 0) {
            html += `
                <div class="alert">
                    <div class="alert-title">⚠️ Security Alerts</div>
                    <ul class="alert-list">
                        ${ssl.alerts.map(alert => `<li>${escapeHtml(alert)}</li>`).join('')}
                    </ul>
                </div>
            `;
        }
        
        // Recommendations
        if (data.recommendations && data.recommendations.length > 0) {
            html += `
                <div class="recommendation">
                    <div class="recommendation-title">💡 Recommendations</div>
                    <ul class="recommendation-list">
                        ${data.recommendations.map(rec => `<li>${escapeHtml(rec)}</li>`).join('')}
                    </ul>
                </div>
            `;
        }
    }
    
    // Server Information
    html += `
        <div class="result-section">
            <h3>🖥️ Server Information</h3>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">IP Address</div>
                    <div class="info-value">${escapeHtml(data.ip || 'N/A')}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">Port</div>
                    <div class="info-value">${data.port}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">Server</div>
                    <div class="info-value">${escapeHtml(data.server || 'Unknown')}</div>
                </div>
            </div>
        </div>
    `;
    
    // IP Geolocation
    if (ipInfo && ipInfo.query) {
        html += `
            <div class="result-section">
                <h3>🌍 IP Geolocation</h3>
                <div class="info-grid">
                    <div class="info-item">
                        <div class="info-label">Country</div>
                        <div class="info-value">${escapeHtml(ipInfo.country || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Region</div>
                        <div class="info-value">${escapeHtml(ipInfo.regionName || ipInfo.region || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">City</div>
                        <div class="info-value">${escapeHtml(ipInfo.city || 'N/A')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">ISP/Organization</div>
                        <div class="info-value">${escapeHtml(ipInfo.org || ipInfo.isp || 'N/A')}</div>
                    </div>
                </div>
            </div>
        `;
    }
    
    div.innerHTML = html;
    return div;
}

// Display Error
function displayError(message) {
    const resultsDiv = document.getElementById('results');
    const resultsContent = document.getElementById('resultsContent');
    
    resultsContent.innerHTML = `
        <div class="error-message">
            ${escapeHtml(message)}
        </div>
    `;
    
    resultsDiv.style.display = 'block';
    resultsDiv.scrollIntoView({ behavior: 'smooth' });
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    if (text === null || text === undefined) return '';
    const div = document.createElement('div');
    div.textContent = text.toString();
    return div.innerHTML;
}
