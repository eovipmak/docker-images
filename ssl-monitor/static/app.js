// SSL Monitor Application JavaScript

// HTML escape function to prevent XSS
function escapeHtml(unsafe) {
    if (unsafe === null || unsafe === undefined) {
        return '';
    }
    return String(unsafe)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

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
                // Create stats container using safe DOM methods
                const container = document.createElement('div');
                container.className = 'row';
                
                const statsData = [
                    { label: 'Total Checks', value: stats.total_checks, color: '' },
                    { label: 'Successful', value: stats.successful_checks, color: 'text-success' },
                    { label: 'Errors', value: stats.error_checks, color: 'text-danger' },
                    { label: 'Unique Domains', value: stats.unique_domains, color: 'text-primary' }
                ];
                
                statsData.forEach(stat => {
                    const col = document.createElement('div');
                    col.className = 'col-md-3 col-sm-6';
                    
                    const statsItem = document.createElement('div');
                    statsItem.className = 'stats-item';
                    
                    const statsNumber = document.createElement('div');
                    statsNumber.className = `stats-number ${stat.color}`;
                    statsNumber.textContent = stat.value;
                    
                    const statsLabel = document.createElement('div');
                    statsLabel.className = 'stats-label';
                    statsLabel.textContent = stat.label;
                    
                    statsItem.appendChild(statsNumber);
                    statsItem.appendChild(statsLabel);
                    col.appendChild(statsItem);
                    container.appendChild(col);
                });
                
                event.detail.target.innerHTML = '';
                event.detail.target.appendChild(container);
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
            event.detail.target.innerHTML = '';
            event.detail.target.appendChild(renderCheckResult(response));
        } catch (e) {
            console.error('Error parsing check result:', e);
            const alert = document.createElement('div');
            alert.className = 'alert alert-danger';
            alert.innerHTML = '<i class="bi bi-exclamation-triangle-fill"></i> Error processing response';
            event.detail.target.innerHTML = '';
            event.detail.target.appendChild(alert);
        }
        event.detail.shouldSwap = false;
    }
    
    // Handle history response
    if (event.detail.target.id === 'history-list') {
        try {
            const response = JSON.parse(event.detail.xhr.responseText);
            if (response.status === 'success') {
                event.detail.target.innerHTML = '';
                const historyContainer = renderHistory(response.history);
                event.detail.target.appendChild(historyContainer);
            }
        } catch (e) {
            console.error('Error parsing history:', e);
        }
        event.detail.shouldSwap = false;
    }
});

function renderCheckResult(response) {
    const card = document.createElement('div');
    card.className = 'card';
    
    if (response.status === 'error') {
        const alert = document.createElement('div');
        alert.className = 'alert alert-danger';
        
        const heading = document.createElement('h5');
        heading.innerHTML = '<i class="bi bi-x-circle-fill"></i> Check Failed';
        
        const para = document.createElement('p');
        para.className = 'mb-0';
        para.textContent = response.error || 'Unknown error occurred';
        
        alert.appendChild(heading);
        alert.appendChild(para);
        return alert;
    }
    
    const data = response.data;
    const ssl = data.ssl || {};
    
    // Card header
    const cardHeader = document.createElement('div');
    cardHeader.className = 'card-header bg-success text-white';
    const headerTitle = document.createElement('h5');
    headerTitle.className = 'mb-0';
    headerTitle.innerHTML = '<i class="bi bi-check-circle-fill"></i> Check Result';
    cardHeader.appendChild(headerTitle);
    
    // Card body
    const cardBody = document.createElement('div');
    cardBody.className = 'card-body';
    
    // Header section with domain/IP
    const headerDiv = document.createElement('div');
    headerDiv.className = 'd-flex justify-content-between align-items-start mb-3';
    
    const infoDiv = document.createElement('div');
    const domainHeading = document.createElement('h5');
    domainHeading.className = 'domain-text';
    domainHeading.textContent = data.domain || data.ip;
    
    const ipInfo = document.createElement('small');
    ipInfo.className = 'text-muted';
    ipInfo.textContent = `IP: ${data.ip} | Port: ${data.port}`;
    
    infoDiv.appendChild(domainHeading);
    infoDiv.appendChild(ipInfo);
    
    // Status badge
    const statusBadge = document.createElement('span');
    statusBadge.className = 'badge';
    if (data.sslStatus === 'success') {
        statusBadge.className += ' bg-success';
        statusBadge.textContent = 'Valid SSL';
    } else if (data.sslStatus === 'warning') {
        statusBadge.className += ' bg-warning';
        statusBadge.textContent = 'Warning';
    } else {
        statusBadge.className += ' bg-danger';
        statusBadge.textContent = 'Error';
    }
    
    headerDiv.appendChild(infoDiv);
    headerDiv.appendChild(statusBadge);
    cardBody.appendChild(headerDiv);
    
    // SSL Info table
    const table = document.createElement('table');
    table.className = 'table table-sm ssl-info-table';
    const tbody = document.createElement('tbody');
    
    const tableData = [
        ['Common Name', ssl.subject?.commonName || 'N/A'],
        ['Issuer', ssl.issuer?.commonName || 'N/A'],
        ['Valid From', ssl.notBefore || 'N/A'],
        ['Valid Until', ssl.notAfter || 'N/A'],
        ['TLS Version', ssl.tlsVersion || 'N/A'],
        ['Server', data.server || 'Unknown'],
        ['Location', `${data.ip_info?.city || ''} ${data.ip_info?.region || ''} ${data.ip_info?.country || 'N/A'}`.trim()]
    ];
    
    tableData.forEach(([label, value]) => {
        const tr = document.createElement('tr');
        const th = document.createElement('th');
        th.textContent = label;
        const td = document.createElement('td');
        td.textContent = value;
        tr.appendChild(th);
        tr.appendChild(td);
        tbody.appendChild(tr);
    });
    
    // Days until expiration row with badge
    if (ssl.daysUntilExpiration !== undefined) {
        const tr = document.createElement('tr');
        const th = document.createElement('th');
        th.textContent = 'Days Until Expiration';
        const td = document.createElement('td');
        const badge = document.createElement('span');
        badge.className = `badge ${ssl.daysUntilExpiration < 30 ? 'bg-danger' : 'bg-success'}`;
        badge.textContent = `${ssl.daysUntilExpiration} days`;
        td.appendChild(badge);
        tr.appendChild(th);
        tr.appendChild(td);
        tbody.appendChild(tr);
    }
    
    table.appendChild(tbody);
    cardBody.appendChild(table);
    
    // Build alerts section
    if (ssl.alerts && ssl.alerts.length > 0) {
        const alertsDiv = document.createElement('div');
        alertsDiv.className = 'mt-3';
        const alertsHeading = document.createElement('h6');
        alertsHeading.innerHTML = '<i class="bi bi-exclamation-triangle-fill text-warning"></i> Security Alerts';
        alertsDiv.appendChild(alertsHeading);
        
        ssl.alerts.forEach(alert => {
            const alertDetail = document.createElement('div');
            alertDetail.className = 'alert-detail';
            alertDetail.textContent = alert;
            alertsDiv.appendChild(alertDetail);
        });
        
        cardBody.appendChild(alertsDiv);
    }
    
    // Build recommendations section
    if (data.recommendations && data.recommendations.length > 0) {
        const recDiv = document.createElement('div');
        recDiv.className = 'mt-3';
        const recHeading = document.createElement('h6');
        recHeading.innerHTML = '<i class="bi bi-lightbulb-fill text-info"></i> Recommendations';
        recDiv.appendChild(recHeading);
        
        data.recommendations.forEach(rec => {
            const recItem = document.createElement('div');
            recItem.className = 'recommendation-item';
            recItem.textContent = rec;
            recDiv.appendChild(recItem);
        });
        
        cardBody.appendChild(recDiv);
    }
    
    // Timestamp
    const timestamp = document.createElement('small');
    timestamp.className = 'text-muted';
    timestamp.innerHTML = `<i class="bi bi-clock"></i> Checked at: ${escapeHtml(new Date(data.checkedAt).toLocaleString())}`;
    cardBody.appendChild(timestamp);
    
    card.appendChild(cardHeader);
    card.appendChild(cardBody);
    
    return card;
}

function renderHistory(history) {
    const container = document.createDocumentFragment();
    
    if (!history || history.length === 0) {
        const emptyDiv = document.createElement('div');
        emptyDiv.className = 'text-center p-4 text-muted';
        emptyDiv.innerHTML = '<i class="bi bi-inbox" style="font-size: 3rem;"></i>';
        
        const emptyText = document.createElement('p');
        emptyText.className = 'mt-2';
        emptyText.textContent = 'No check history yet';
        
        emptyDiv.appendChild(emptyText);
        container.appendChild(emptyDiv);
        return container;
    }
    
    history.forEach(item => {
        const historyItem = document.createElement('div');
        historyItem.className = 'history-item';
        
        // Determine status icon and class
        let statusIcon = '';
        if (item.status === 'success') {
            statusIcon = '<i class="bi bi-check-circle-fill text-success"></i>';
            historyItem.className += ' border-start border-success border-3';
        } else if (item.status === 'warning') {
            statusIcon = '<i class="bi bi-exclamation-triangle-fill text-warning"></i>';
            historyItem.className += ' border-start border-warning border-3';
        } else {
            statusIcon = '<i class="bi bi-x-circle-fill text-danger"></i>';
            historyItem.className += ' border-start border-danger border-3';
        }
        
        const flexContainer = document.createElement('div');
        flexContainer.className = 'd-flex justify-content-between align-items-start';
        
        // Left side - domain/IP info
        const leftDiv = document.createElement('div');
        leftDiv.className = 'flex-grow-1';
        
        const topRow = document.createElement('div');
        topRow.className = 'd-flex align-items-center gap-2';
        topRow.innerHTML = statusIcon;
        
        const domainSpan = document.createElement('span');
        domainSpan.className = 'domain-text';
        domainSpan.textContent = item.domain || item.ip;
        
        const portBadge = document.createElement('span');
        portBadge.className = 'badge bg-secondary';
        portBadge.textContent = item.port;
        
        topRow.appendChild(domainSpan);
        topRow.appendChild(portBadge);
        
        const timestampDiv = document.createElement('div');
        timestampDiv.className = 'timestamp-text mt-1';
        timestampDiv.innerHTML = '<i class="bi bi-clock"></i> ';
        const timestampText = document.createTextNode(new Date(item.checked_at).toLocaleString());
        timestampDiv.appendChild(timestampText);
        
        leftDiv.appendChild(topRow);
        leftDiv.appendChild(timestampDiv);
        
        // Right side - SSL status
        const rightDiv = document.createElement('div');
        rightDiv.className = 'text-end';
        
        const sslBadge = document.createElement('span');
        sslBadge.className = `badge status-badge ${item.ssl_status === 'success' ? 'bg-success' : 'bg-danger'}`;
        sslBadge.textContent = `SSL: ${item.ssl_status}`;
        
        rightDiv.appendChild(sslBadge);
        
        flexContainer.appendChild(leftDiv);
        flexContainer.appendChild(rightDiv);
        historyItem.appendChild(flexContainer);
        
        container.appendChild(historyItem);
    });
    
    return container;
}

// Handle HTMX errors
document.body.addEventListener('htmx:responseError', function(event) {
    console.error('HTMX Error:', event.detail);
    
    if (event.detail.target.id === 'check-result') {
        const alert = document.createElement('div');
        alert.className = 'alert alert-danger';
        
        const heading = document.createElement('h5');
        heading.innerHTML = '<i class="bi bi-exclamation-triangle-fill"></i> Error';
        
        const para = document.createElement('p');
        para.className = 'mb-0';
        para.textContent = 'Failed to check SSL certificate. The SSL checker service may be unavailable.';
        
        alert.appendChild(heading);
        alert.appendChild(para);
        
        event.detail.target.innerHTML = '';
        event.detail.target.appendChild(alert);
    }
});
