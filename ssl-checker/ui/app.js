// Single Check Form Handler
document.getElementById('singleCheckForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const form = e.target;
    const button = form.querySelector('button[type="submit"]');
    const btnText = button.querySelector('.btn-text');
    const spinner = button.querySelector('.spinner');
    
    const domain = form.domain.value.trim();
    const ip = form.ip.value.trim();
    const port = form.port.value;
    
    // Validation
    if (!domain && !ip) {
        alert('Please provide either a domain name or IP address');
        return;
    }
    
    if (domain && ip) {
        alert('Please provide only domain OR IP, not both');
        return;
    }
    
    // Show loading state
    button.disabled = true;
    btnText.style.display = 'none';
    spinner.style.display = 'inline-block';
    
    try {
        const params = new URLSearchParams({ port });
        if (domain) params.append('domain', domain);
        if (ip) params.append('ip', ip);
        
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

// Batch Check Form Handler
document.getElementById('batchCheckForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const form = e.target;
    const button = form.querySelector('button[type="submit"]');
    const btnText = button.querySelector('.btn-text');
    const spinner = button.querySelector('.spinner');
    
    const domainsText = form.domains.value.trim();
    const ipsText = form.ips.value.trim();
    const port = parseInt(form.port.value);
    
    // Parse domains and IPs
    const domains = domainsText ? domainsText.split('\n').map(d => d.trim()).filter(d => d) : [];
    const ips = ipsText ? ipsText.split('\n').map(i => i.trim()).filter(i => i) : [];
    
    if (domains.length === 0 && ips.length === 0) {
        alert('Please provide at least one domain or IP address');
        return;
    }
    
    // Show loading state
    button.disabled = true;
    btnText.style.display = 'none';
    spinner.style.display = 'inline-block';
    
    try {
        const response = await fetch('/api/batch_check', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ domains, ips, port })
        });
        
        const data = await response.json();
        
        if (data.status === 'success' && data.results) {
            displayResults(data.results);
        } else {
            displayError('Failed to check SSL certificates');
        }
    } catch (error) {
        console.error('Error:', error);
        displayError('Failed to check SSL certificates. Please try again.');
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
                <div class="info-grid">
                    <div class="info-item">
                        <div class="info-label">Subject CN</div>
                        <div class="info-value">${escapeHtml(ssl.subject?.commonName || 'N/A')}</div>
                    </div>
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
