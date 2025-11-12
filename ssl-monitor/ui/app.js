/**
 * SSL Monitor Frontend - Vanilla JavaScript
 */

// API endpoints
const API_BASE = '/api';

/**
 * Load statistics
 */
async function loadStats() {
    try {
        const response = await fetch(`${API_BASE}/stats`);
        const result = await response.json();
        
        if (result.status === 'success') {
            document.getElementById('stat-total').textContent = result.data.total_checks;
            document.getElementById('stat-valid').textContent = result.data.valid_certificates;
            document.getElementById('stat-invalid').textContent = result.data.invalid_certificates;
            document.getElementById('stat-rate').textContent = result.data.success_rate + '%';
        }
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

/**
 * Load check history
 */
async function loadHistory() {
    try {
        const response = await fetch(`${API_BASE}/history?limit=20`);
        const result = await response.json();
        
        if (result.status === 'success') {
            renderHistory(result.data);
        }
    } catch (error) {
        console.error('Error loading history:', error);
        document.getElementById('historyTableBody').innerHTML = 
            '<tr><td colspan="6" class="text-center text-danger">Error loading history</td></tr>';
    }
}

/**
 * Render history table
 */
function renderHistory(checks) {
    const tbody = document.getElementById('historyTableBody');
    
    if (checks.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted">No checks yet. Start by checking a domain!</td></tr>';
        return;
    }
    
    tbody.innerHTML = checks.map(check => {
        const statusBadge = check.ssl_valid 
            ? '<span class="badge bg-success"><i class="bi bi-check-circle"></i> Valid</span>'
            : '<span class="badge bg-danger"><i class="bi bi-x-circle"></i> Invalid</span>';
        
        const date = new Date(check.checked_at);
        const formattedDate = date.toLocaleString();
        
        return `
            <tr>
                <td><strong>${check.domain || check.ip}</strong></td>
                <td>${check.port}</td>
                <td>${statusBadge}</td>
                <td>${check.server_info || '-'}</td>
                <td><small>${formattedDate}</small></td>
                <td>
                    <button class="btn btn-sm btn-outline-primary" onclick="viewDetails(${check.id})">
                        <i class="bi bi-eye"></i>
                    </button>
                </td>
            </tr>
        `;
    }).join('');
}

/**
 * Check SSL certificate
 */
async function checkCertificate(event) {
    event.preventDefault();
    
    const target = document.getElementById('target').value.trim();
    const port = document.getElementById('port').value;
    const resultsDiv = document.getElementById('checkResults');
    const form = event.target;
    
    // Add loading state
    form.classList.add('htmx-request');
    
    // Create loading message safely to prevent XSS
    const alertDiv = document.createElement('div');
    alertDiv.className = 'alert alert-info mt-3';
    const spinner = document.createElement('div');
    spinner.className = 'spinner-border spinner-border-sm me-2';
    spinner.setAttribute('role', 'status');
    alertDiv.appendChild(spinner);
    alertDiv.appendChild(document.createTextNode(`Checking certificate for ${target}:${port}...`));
    resultsDiv.innerHTML = '';
    resultsDiv.appendChild(alertDiv);
    
    try {
        // Determine if it's an IP or domain
        const isIP = /^(\d{1,3}\.){3}\d{1,3}$/.test(target) || target.includes(':');
        const params = new URLSearchParams({
            [isIP ? 'ip' : 'domain']: target,
            port: port
        });
        
        const response = await fetch(`${API_BASE}/check?${params}`);
        const result = await response.json();
        
        if (result.status === 'success') {
            displayCheckResult(result);
            // Reload stats and history
            loadStats();
            loadHistory();
        } else {
            resultsDiv.innerHTML = `
                <div class="alert alert-danger mt-3">
                    <i class="bi bi-exclamation-triangle"></i>
                    <strong>Error:</strong> ${result.error || 'Failed to check certificate'}
                </div>
            `;
        }
    } catch (error) {
        resultsDiv.innerHTML = `
            <div class="alert alert-danger mt-3">
                <i class="bi bi-exclamation-triangle"></i>
                <strong>Error:</strong> ${error.message}
            </div>
        `;
    } finally {
        form.classList.remove('htmx-request');
    }
}

/**
 * Display check result
 */
function displayCheckResult(result) {
    const resultsDiv = document.getElementById('checkResults');
    const data = result.data;
    const ssl = data.ssl || {};
    
    const isValid = data.sslStatus === 'success';
    const alertClass = isValid ? 'alert-success' : 'alert-danger';
    const icon = isValid ? 'bi-check-circle-fill' : 'bi-x-circle-fill';
    
    let html = `
        <div class="alert ${alertClass} mt-3">
            <h6><i class="bi ${icon}"></i> ${isValid ? 'Certificate Valid' : 'Certificate Invalid'}</h6>
            <hr>
            <p class="mb-0"><strong>Domain:</strong> ${data.domain || data.ip}</p>
            <p class="mb-0"><strong>Port:</strong> ${data.port}</p>
    `;
    
    if (isValid && ssl) {
        html += `
            <p class="mb-0"><strong>Issuer:</strong> ${ssl.issuer?.commonName || 'N/A'}</p>
            <p class="mb-0"><strong>Valid Until:</strong> ${ssl.notAfter || 'N/A'}</p>
            <p class="mb-0"><strong>Days Until Expiration:</strong> ${ssl.daysUntilExpiration || 'N/A'}</p>
            <p class="mb-0"><strong>TLS Version:</strong> ${ssl.tlsVersion || 'N/A'}</p>
        `;
    }
    
    html += '</div>';
    
    // Alerts
    if (ssl.alerts && ssl.alerts.length > 0) {
        html += '<div class="alert alert-warning mt-2"><strong>Alerts:</strong><ul class="mb-0 mt-2">';
        ssl.alerts.forEach(alert => {
            html += `<li>${alert}</li>`;
        });
        html += '</ul></div>';
    }
    
    // Recommendations
    if (data.recommendations && data.recommendations.length > 0) {
        html += '<div class="alert alert-info mt-2"><strong>Recommendations:</strong><ul class="mb-0 mt-2">';
        data.recommendations.forEach(rec => {
            html += `<li>${rec}</li>`;
        });
        html += '</ul></div>';
    }
    
    resultsDiv.innerHTML = html;
}

/**
 * View check details
 */
async function viewDetails(checkId) {
    const modal = new bootstrap.Modal(document.getElementById('detailModal'));
    const modalBody = document.getElementById('detailModalBody');
    
    modalBody.innerHTML = `
        <div class="text-center">
            <div class="spinner-border text-primary" role="status"></div>
        </div>
    `;
    
    modal.show();
    
    try {
        const response = await fetch(`${API_BASE}/history/${checkId}`);
        const result = await response.json();
        
        if (result.status === 'success') {
            renderDetailModal(result.data);
        } else {
            modalBody.innerHTML = `
                <div class="alert alert-danger">
                    Error loading details: ${result.error}
                </div>
            `;
        }
    } catch (error) {
        modalBody.innerHTML = `
            <div class="alert alert-danger">
                Error: ${error.message}
            </div>
        `;
    }
}

/**
 * Render detail modal content
 */
function renderDetailModal(data) {
    const modalBody = document.getElementById('detailModalBody');
    const cert = data.certificate_info || {};
    const ipInfo = data.ip_info || {};
    
    let html = '<div class="row">';
    
    // Basic Info
    html += `
        <div class="col-md-6">
            <h6 class="border-bottom pb-2">Basic Information</h6>
            <table class="table table-sm">
                <tr><th>Domain:</th><td>${data.domain || '-'}</td></tr>
                <tr><th>IP Address:</th><td>${data.ip || '-'}</td></tr>
                <tr><th>Port:</th><td>${data.port}</td></tr>
                <tr><th>Server:</th><td>${data.server_info || '-'}</td></tr>
                <tr><th>Valid:</th><td>
                    ${data.ssl_valid 
                        ? '<span class="badge bg-success">Yes</span>' 
                        : '<span class="badge bg-danger">No</span>'}
                </td></tr>
                <tr><th>Checked:</th><td>${new Date(data.checked_at).toLocaleString()}</td></tr>
            </table>
        </div>
    `;
    
    // Certificate Info
    html += `
        <div class="col-md-6">
            <h6 class="border-bottom pb-2">Certificate Details</h6>
            <table class="table table-sm">
                <tr><th>Subject:</th><td>${cert.subject?.commonName || '-'}</td></tr>
                <tr><th>Issuer:</th><td>${cert.issuer?.commonName || '-'}</td></tr>
                <tr><th>Valid From:</th><td>${cert.notBefore || '-'}</td></tr>
                <tr><th>Valid Until:</th><td>${cert.notAfter || '-'}</td></tr>
                <tr><th>Days Until Expiry:</th><td>${cert.daysUntilExpiration || '-'}</td></tr>
                <tr><th>TLS Version:</th><td>${cert.tlsVersion || '-'}</td></tr>
                <tr><th>Cipher Suite:</th><td><small>${cert.cipherSuite || '-'}</small></td></tr>
            </table>
        </div>
    `;
    
    html += '</div>';
    
    // IP Geolocation
    if (ipInfo.country) {
        html += `
            <div class="mt-3">
                <h6 class="border-bottom pb-2">IP Geolocation</h6>
                <div class="row">
                    <div class="col-md-6">
                        <p><strong>Country:</strong> ${ipInfo.country || '-'}</p>
                        <p><strong>Region:</strong> ${ipInfo.region || '-'}</p>
                        <p><strong>City:</strong> ${ipInfo.city || '-'}</p>
                    </div>
                    <div class="col-md-6">
                        <p><strong>Organization:</strong> ${ipInfo.org || '-'}</p>
                        <p><strong>Timezone:</strong> ${ipInfo.timezone || '-'}</p>
                    </div>
                </div>
            </div>
        `;
    }
    
    // Alerts
    if (data.alerts && data.alerts.length > 0) {
        html += `
            <div class="mt-3">
                <h6 class="border-bottom pb-2 text-warning">Alerts</h6>
                <ul class="list-group">
        `;
        data.alerts.forEach(alert => {
            html += `<li class="list-group-item list-group-item-warning">${alert}</li>`;
        });
        html += '</ul></div>';
    }
    
    // Recommendations
    if (data.recommendations && data.recommendations.length > 0) {
        html += `
            <div class="mt-3">
                <h6 class="border-bottom pb-2 text-info">Recommendations</h6>
                <ul class="list-group">
        `;
        data.recommendations.forEach(rec => {
            html += `<li class="list-group-item list-group-item-info">${rec}</li>`;
        });
        html += '</ul></div>';
    }
    
    // Error message
    if (data.error_message) {
        html += `
            <div class="mt-3">
                <div class="alert alert-danger">
                    <strong>Error:</strong> ${data.error_message}
                </div>
            </div>
        `;
    }
    
    modalBody.innerHTML = html;
}
