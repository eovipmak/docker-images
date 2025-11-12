// Initialize language on page load
document.addEventListener('DOMContentLoaded', () => {
    initLanguage();
    
    // Setup language toggle button
    const langToggle = document.getElementById('languageToggle');
    if (langToggle) {
        langToggle.addEventListener('click', () => {
            const newLang = getCurrentLanguage() === 'vi' ? 'en' : 'vi';
            setLanguage(newLang);
        });
    }
});

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
        alert(t('provideDomain'));
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
        alert(t('invalidTarget'));
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
        
        // Store results for language switching
        window.lastResults = [data];
        displayResults([data]);
    } catch (error) {
        console.error('Error:', error);
        displayError(t('checkFailed'));
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
    
    // Update results title
    const resultsTitle = document.querySelector('#results .card h2');
    if (resultsTitle) {
        resultsTitle.textContent = t('resultsTitle');
    }
    
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
                <div class="result-title">${t('result')} ${index + 1}</div>
                <span class="status-badge status-error">${t('statusError')}</span>
            </div>
            <div class="error-message">
                ${escapeHtml(result.error || t('errorOccurred'))}
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
            <span class="status-badge status-${data.sslStatus}">${t('status' + capitalizeFirst(data.sslStatus))}</span>
        </div>
    `;
    
    // SSL Information
    if (ssl) {
        html += `
            <div class="result-section">
                <h3>${t('sslCertificate')}</h3>
                <div class="info-grid">`;
        
        // Display all subject fields
        if (ssl.subject) {
            const subjectEntries = Object.entries(ssl.subject);
            if (subjectEntries.length > 0) {
                subjectEntries.forEach(([key, value]) => {
                    let label = key;
                    if (key === 'commonName') label = t('subjectCN');
                    else if (key === 'organizationName') label = t('subjectOrganization');
                    else if (key === 'organizationalUnitName') label = t('subjectOrgUnit');
                    else if (key === 'countryName') label = t('subjectCountry');
                    else if (key === 'stateOrProvinceName') label = t('subjectState');
                    else if (key === 'localityName') label = t('subjectLocality');
                    else label = `Subject ${key}`;
                    
                    html += `
                    <div class="info-item">
                        <div class="info-label">${escapeHtml(label)}</div>
                        <div class="info-value">${escapeHtml(value || t('notAvailable'))}</div>
                    </div>`;
                });
            }
        }
        
        // Issuer information
        if (ssl.issuer) {
            html += `
                <div class="info-item">
                    <div class="info-label">${t('issuer')}</div>
                    <div class="info-value">${escapeHtml(ssl.issuer.commonName || t('notAvailable'))}</div>
                </div>`;
            
            if (ssl.issuer.organizationName) {
                html += `
                <div class="info-item">
                    <div class="info-label">${t('issuerOrg')}</div>
                    <div class="info-value">${escapeHtml(ssl.issuer.organizationName)}</div>
                </div>`;
            }
            
            if (ssl.issuer.countryName) {
                html += `
                <div class="info-item">
                    <div class="info-label">${t('issuerCountry')}</div>
                    <div class="info-value">${escapeHtml(ssl.issuer.countryName)}</div>
                </div>`;
            }
        }
        
        // Certificate details
        html += `
                    <div class="info-item">
                        <div class="info-label">${t('version')}</div>
                        <div class="info-value">${ssl.version !== null && ssl.version !== undefined ? ssl.version : t('notAvailable')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('serialNumber')}</div>
                        <div class="info-value">${escapeHtml(ssl.serialNumber || t('notAvailable'))}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('validFrom')}</div>
                        <div class="info-value">${escapeHtml(ssl.notBefore || t('notAvailable'))}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('validUntil')}</div>
                        <div class="info-value">${escapeHtml(ssl.notAfter || t('notAvailable'))}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('daysUntilExpiration')}</div>
                        <div class="info-value">${ssl.daysUntilExpiration !== null ? ssl.daysUntilExpiration : t('notAvailable')}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('signatureAlgorithm')}</div>
                        <div class="info-value">${escapeHtml(ssl.signatureAlgorithm || t('notAvailable'))}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('tlsVersion')}</div>
                        <div class="info-value">${escapeHtml(ssl.tlsVersion || t('notAvailable'))}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">${t('cipherSuite')}</div>
                        <div class="info-value">${escapeHtml(ssl.cipherSuite || t('notAvailable'))}</div>
                    </div>`;
        
        // Subject Alternative Names
        if (ssl.subjectAltNames && ssl.subjectAltNames.length > 0) {
            const sanList = ssl.subjectAltNames.map(san => {
                if (Array.isArray(san)) {
                    return `${san[0]}: ${san[1]}`;
                }
                return san;
            }).join(', ');
            
            html += `
                    <div class="info-item" style="grid-column: 1 / -1;">
                        <div class="info-label">${t('subjectAltNames')}</div>
                        <div class="info-value">${escapeHtml(sanList)}</div>
                    </div>`;
        }
        
        html += `
                </div>
            </div>
        `;
        
        // Alerts
        if (ssl.alerts && ssl.alerts.length > 0) {
            html += `
                <div class="alert">
                    <div class="alert-title">${t('securityAlerts')}</div>
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
                    <div class="recommendation-title">${t('recommendations')}</div>
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
            <h3>${t('serverInformation')}</h3>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">${t('ipAddress')}</div>
                    <div class="info-value">${escapeHtml(data.ip || t('notAvailable'))}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">${t('port')}</div>
                    <div class="info-value">${data.port}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">${t('server')}</div>
                    <div class="info-value">${escapeHtml(data.server || t('unknown'))}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">${t('sslStatus')}</div>
                    <div class="info-value">${escapeHtml(data.sslStatus || t('notAvailable'))}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">${t('serverStatus')}</div>
                    <div class="info-value">${escapeHtml(data.serverStatus || t('notAvailable'))}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">${t('ipStatus')}</div>
                    <div class="info-value">${escapeHtml(data.ipStatus || t('notAvailable'))}</div>
                </div>`;
    
    if (data.sslErrorType) {
        html += `
                <div class="info-item">
                    <div class="info-label">${t('errorType')}</div>
                    <div class="info-value">${escapeHtml(data.sslErrorType)}</div>
                </div>`;
    }
    
    if (data.checkedAt) {
        html += `
                <div class="info-item">
                    <div class="info-label">${t('checkedAt')}</div>
                    <div class="info-value">${escapeHtml(data.checkedAt)}</div>
                </div>`;
    }
    
    html += `
            </div>
        </div>
    `;
    
    // IP Geolocation - Display all fields from the reference sample
    if (ipInfo && ipInfo.query) {
        html += `
            <div class="result-section">
                <h3>${t('ipGeolocation')}</h3>
                <div class="info-grid">`;
        
        // Display all IP info fields
        const ipInfoFields = [
            { key: 'query', label: t('ipAddress') },
            { key: 'continent', label: t('continent') },
            { key: 'continentCode', label: t('continentCode') },
            { key: 'country', label: t('country') },
            { key: 'countryCode', label: t('countryCode') },
            { key: 'region', label: t('region') },
            { key: 'regionName', label: t('regionName') },
            { key: 'city', label: t('city') },
            { key: 'district', label: t('district') },
            { key: 'zip', label: t('zip') },
            { key: 'isp', label: t('isp') },
            { key: 'org', label: t('org') },
            { key: 'as', label: t('asn') },
            { key: 'asname', label: t('asname') },
            { key: 'reverse', label: t('reverse') }
        ];
        
        ipInfoFields.forEach(field => {
            const value = ipInfo[field.key];
            if (value !== null && value !== undefined && value !== '') {
                html += `
                    <div class="info-item">
                        <div class="info-label">${field.label}</div>
                        <div class="info-value">${escapeHtml(value)}</div>
                    </div>`;
            }
        });
        
        // Coordinates
        if (ipInfo.lat !== null && ipInfo.lat !== undefined && ipInfo.lon !== null && ipInfo.lon !== undefined) {
            html += `
                    <div class="info-item">
                        <div class="info-label">${t('coordinates')}</div>
                        <div class="info-value">${ipInfo.lat}, ${ipInfo.lon}</div>
                    </div>`;
        }
        
        // Boolean fields
        const boolFields = [
            { key: 'mobile', label: t('mobile') },
            { key: 'proxy', label: t('proxy') },
            { key: 'hosting', label: t('hosting') }
        ];
        
        boolFields.forEach(field => {
            if (ipInfo[field.key] !== null && ipInfo[field.key] !== undefined) {
                html += `
                    <div class="info-item">
                        <div class="info-label">${field.label}</div>
                        <div class="info-value">${ipInfo[field.key] ? t('yes') : t('no')}</div>
                    </div>`;
            }
        });
        
        html += `
                </div>
            </div>
        `;
    }
    
    div.innerHTML = html;
    return div;
}

// Helper function to capitalize first letter
function capitalizeFirst(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
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