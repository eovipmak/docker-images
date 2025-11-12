// SSL Monitor Application JavaScript

// HTMX event listeners
document.body.addEventListener('htmx:afterSwap', function(event) {
    // If it's a check result, trigger history refresh
    if (event.detail.target.id === 'check-result') {
        // Dispatch custom event to refresh history
        document.body.dispatchEvent(new CustomEvent('checkCompleted'));
    }
});

// Custom response transformers for better UI
document.body.addEventListener('htmx:beforeSwap', function(event) {
    // Handle stats response
    if (event.detail.target.id === 'stats-container') {
        try {
            const response = JSON.parse(event.detail.xhr.responseText);
            if (response.status === 'success') {
                const stats = response.stats;
                event.detail.target.innerHTML = `
                    <div class="row">
                        <div class="col-md-3 col-sm-6">
                            <div class="stats-item">
                                <div class="stats-number">${stats.total_checks}</div>
                                <div class="stats-label">Total Checks</div>
                            </div>
                        </div>
                        <div class="col-md-3 col-sm-6">
                            <div class="stats-item">
                                <div class="stats-number text-success">${stats.successful_checks}</div>
                                <div class="stats-label">Successful</div>
                            </div>
                        </div>
                        <div class="col-md-3 col-sm-6">
                            <div class="stats-item">
                                <div class="stats-number text-danger">${stats.error_checks}</div>
                                <div class="stats-label">Errors</div>
                            </div>
                        </div>
                        <div class="col-md-3 col-sm-6">
                            <div class="stats-item">
                                <div class="stats-number text-primary">${stats.unique_domains}</div>
                                <div class="stats-label">Unique Domains</div>
                            </div>
                        </div>
                    </div>
                `;
            }
        } catch (e) {
            console.error('Error parsing stats:', e);
        }
        event.detail.shouldSwap = false;
    }
    
    // Handle check result response
    if (event.detail.target.id === 'check-result') {
        try {
            const response = JSON.parse(event.detail.xhr.responseText);
            event.detail.target.innerHTML = renderCheckResult(response);
        } catch (e) {
            console.error('Error parsing check result:', e);
            event.detail.target.innerHTML = `
                <div class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle-fill"></i> Error processing response
                </div>
            `;
        }
        event.detail.shouldSwap = false;
    }
    
    // Handle history response
    if (event.detail.target.id === 'history-list') {
        try {
            const response = JSON.parse(event.detail.xhr.responseText);
            if (response.status === 'success') {
                event.detail.target.innerHTML = renderHistory(response.history);
            }
        } catch (e) {
            console.error('Error parsing history:', e);
        }
        event.detail.shouldSwap = false;
    }
});

function renderCheckResult(response) {
    if (response.status === 'error') {
        return `
            <div class="alert alert-danger">
                <h5><i class="bi bi-x-circle-fill"></i> Check Failed</h5>
                <p class="mb-0">${response.error || 'Unknown error occurred'}</p>
            </div>
        `;
    }
    
    const data = response.data;
    const ssl = data.ssl || {};
    
    // Status badge based on SSL status
    let statusBadge = '';
    if (data.sslStatus === 'success') {
        statusBadge = '<span class="badge bg-success">Valid SSL</span>';
    } else if (data.sslStatus === 'warning') {
        statusBadge = '<span class="badge bg-warning">Warning</span>';
    } else {
        statusBadge = '<span class="badge bg-danger">Error</span>';
    }
    
    // Build alerts section
    let alertsHtml = '';
    if (ssl.alerts && ssl.alerts.length > 0) {
        alertsHtml = `
            <div class="mt-3">
                <h6><i class="bi bi-exclamation-triangle-fill text-warning"></i> Security Alerts</h6>
                ${ssl.alerts.map(alert => `<div class="alert-detail">${alert}</div>`).join('')}
            </div>
        `;
    }
    
    // Build recommendations section
    let recommendationsHtml = '';
    if (data.recommendations && data.recommendations.length > 0) {
        recommendationsHtml = `
            <div class="mt-3">
                <h6><i class="bi bi-lightbulb-fill text-info"></i> Recommendations</h6>
                ${data.recommendations.map(rec => `<div class="recommendation-item">${rec}</div>`).join('')}
            </div>
        `;
    }
    
    return `
        <div class="card">
            <div class="card-header bg-success text-white">
                <h5 class="mb-0"><i class="bi bi-check-circle-fill"></i> Check Result</h5>
            </div>
            <div class="card-body">
                <div class="d-flex justify-content-between align-items-start mb-3">
                    <div>
                        <h5 class="domain-text">${data.domain || data.ip}</h5>
                        <small class="text-muted">IP: ${data.ip} | Port: ${data.port}</small>
                    </div>
                    ${statusBadge}
                </div>
                
                <table class="table table-sm ssl-info-table">
                    <tbody>
                        <tr>
                            <th>Common Name</th>
                            <td>${ssl.subject?.commonName || 'N/A'}</td>
                        </tr>
                        <tr>
                            <th>Issuer</th>
                            <td>${ssl.issuer?.commonName || 'N/A'}</td>
                        </tr>
                        <tr>
                            <th>Valid From</th>
                            <td>${ssl.notBefore || 'N/A'}</td>
                        </tr>
                        <tr>
                            <th>Valid Until</th>
                            <td>${ssl.notAfter || 'N/A'}</td>
                        </tr>
                        <tr>
                            <th>Days Until Expiration</th>
                            <td>
                                ${ssl.daysUntilExpiration !== undefined 
                                    ? `<span class="badge ${ssl.daysUntilExpiration < 30 ? 'bg-danger' : 'bg-success'}">${ssl.daysUntilExpiration} days</span>`
                                    : 'N/A'
                                }
                            </td>
                        </tr>
                        <tr>
                            <th>TLS Version</th>
                            <td>${ssl.tlsVersion || 'N/A'}</td>
                        </tr>
                        <tr>
                            <th>Server</th>
                            <td>${data.server || 'Unknown'}</td>
                        </tr>
                        <tr>
                            <th>Location</th>
                            <td>${data.ip_info?.city || ''} ${data.ip_info?.region || ''} ${data.ip_info?.country || 'N/A'}</td>
                        </tr>
                    </tbody>
                </table>
                
                ${alertsHtml}
                ${recommendationsHtml}
                
                <small class="text-muted">
                    <i class="bi bi-clock"></i> Checked at: ${new Date(data.checkedAt).toLocaleString()}
                </small>
            </div>
        </div>
    `;
}

function renderHistory(history) {
    if (!history || history.length === 0) {
        return `
            <div class="text-center p-4 text-muted">
                <i class="bi bi-inbox" style="font-size: 3rem;"></i>
                <p class="mt-2">No check history yet</p>
            </div>
        `;
    }
    
    return history.map(item => {
        let statusIcon = '';
        let statusClass = '';
        
        if (item.status === 'success') {
            statusIcon = '<i class="bi bi-check-circle-fill text-success"></i>';
            statusClass = 'border-start border-success border-3';
        } else if (item.status === 'warning') {
            statusIcon = '<i class="bi bi-exclamation-triangle-fill text-warning"></i>';
            statusClass = 'border-start border-warning border-3';
        } else {
            statusIcon = '<i class="bi bi-x-circle-fill text-danger"></i>';
            statusClass = 'border-start border-danger border-3';
        }
        
        const timestamp = new Date(item.checked_at).toLocaleString();
        
        return `
            <div class="history-item ${statusClass}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="d-flex align-items-center gap-2">
                            ${statusIcon}
                            <span class="domain-text">${item.domain || item.ip}</span>
                            <span class="badge bg-secondary">${item.port}</span>
                        </div>
                        <div class="timestamp-text mt-1">
                            <i class="bi bi-clock"></i> ${timestamp}
                        </div>
                    </div>
                    <div class="text-end">
                        <span class="badge status-badge ${item.ssl_status === 'success' ? 'bg-success' : 'bg-danger'}">
                            SSL: ${item.ssl_status}
                        </span>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

// Handle HTMX errors
document.body.addEventListener('htmx:responseError', function(event) {
    console.error('HTMX Error:', event.detail);
    
    if (event.detail.target.id === 'check-result') {
        event.detail.target.innerHTML = `
            <div class="alert alert-danger">
                <h5><i class="bi bi-exclamation-triangle-fill"></i> Error</h5>
                <p class="mb-0">Failed to check SSL certificate. The SSL checker service may be unavailable.</p>
            </div>
        `;
    }
});
