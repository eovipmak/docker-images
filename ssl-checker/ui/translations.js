// Translation data for the SSL Checker application
const translations = {
    en: {
        // Header
        title: "🔒 SSL Certificate Checker",
        subtitle: "Check SSL certificate information, server details, and IP geolocation",
        
        // Form
        formTitle: "SSL Certificate Check",
        targetLabel: "Domain or IP Address",
        targetPlaceholder: "example.com:443, 93.184.216.34:8443, or [::1]:443",
        helpText: "Enter a domain or IP address (IPv4/IPv6). Optional port in format domain[:port], IP[:port], or [IPv6]:port (default: 443)",
        checkButton: "Check Certificate",
        
        // Results
        resultsTitle: "Results",
        result: "Result",
        
        // Status badges
        statusSuccess: "Success",
        statusError: "Error",
        statusWarning: "Warning",
        
        // Sections
        sslCertificate: "🔒 SSL Certificate",
        serverInformation: "🖥️ Server Information",
        ipGeolocation: "🌍 IP Geolocation",
        securityAlerts: "⚠️ Security Alerts",
        recommendations: "💡 Recommendations",
        
        // SSL Certificate Fields
        subjectCN: "Subject CN (Common Name)",
        subjectOrganization: "Subject Organization",
        subjectOrgUnit: "Subject Organizational Unit",
        subjectCountry: "Subject Country",
        subjectState: "Subject State/Province",
        subjectLocality: "Subject Locality",
        issuer: "Issuer",
        issuerOrg: "Issuer Organization",
        issuerCountry: "Issuer Country",
        version: "Version",
        serialNumber: "Serial Number",
        validFrom: "Valid From",
        validUntil: "Valid Until",
        daysUntilExpiration: "Days Until Expiration",
        tlsVersion: "TLS Version",
        cipherSuite: "Cipher Suite",
        signatureAlgorithm: "Signature Algorithm",
        subjectAltNames: "Subject Alternative Names (SAN)",
        
        // Server Information Fields
        ipAddress: "IP Address",
        port: "Port",
        server: "Server",
        
        // IP Geolocation Fields
        continent: "Continent",
        continentCode: "Continent Code",
        country: "Country",
        countryCode: "Country Code",
        region: "Region",
        regionName: "Region Name",
        city: "City",
        district: "District",
        zip: "Zip Code",
        coordinates: "Coordinates (Lat, Lon)",
        isp: "Internet Service Provider",
        org: "Organization",
        asn: "Autonomous System Number (AS)",
        asname: "AS Name",
        reverse: "Reverse DNS",
        mobile: "Mobile",
        proxy: "Proxy",
        hosting: "Hosting",
        
        // Status Fields
        sslStatus: "SSL Status",
        serverStatus: "Server Status",
        ipStatus: "IP Status",
        errorType: "Error Type",
        checkedAt: "Checked At",
        
        // Common
        yes: "Yes",
        no: "No",
        unknown: "Unknown",
        notAvailable: "N/A",
        
        // Errors
        errorOccurred: "An error occurred",
        provideDomain: "Please provide a domain name or IP address",
        invalidTarget: "Please provide a valid domain name or IP address",
        checkFailed: "Failed to check SSL certificate. Please try again.",
        
        // Footer
        footerText: "🔒 SSL Checker API v2.0.0",
        builtWith: "Built with",
        by: "by"
    }
};

// Current language (default: English)
let currentLanguage = 'en';

// Get translation for a key
function t(key) {
    return translations[currentLanguage][key] || key;
}

// Set language
function setLanguage(lang) {
    if (translations[lang]) {
        currentLanguage = lang;
        localStorage.setItem('preferredLanguage', lang);
        updatePageLanguage();
    }
}

// Get current language
function getCurrentLanguage() {
    return currentLanguage;
}

// Initialize language from localStorage or default
function initLanguage() {
    const savedLang = localStorage.getItem('preferredLanguage');
    if (savedLang && translations[savedLang]) {
        currentLanguage = savedLang;
    }
    updatePageLanguage();
}

// Update all text on the page based on current language
function updatePageLanguage() {
    // Update header
    document.querySelector('header h1').textContent = t('title');
    document.querySelector('.subtitle').textContent = t('subtitle');
    
    // Update form
    document.querySelector('.card h2').textContent = t('formTitle');
    document.querySelector('label[for="target"]').textContent = t('targetLabel');
    document.querySelector('#target').placeholder = t('targetPlaceholder');
    document.querySelector('.help-text').textContent = t('helpText');
    document.querySelector('.btn-text').textContent = t('checkButton');
    
    // Update results title if results are visible
    const resultsTitle = document.querySelector('#results .card h2');
    if (resultsTitle) {
        resultsTitle.textContent = t('resultsTitle');
    }
    
    // Update document lang attribute for accessibility (screen readers, search engines)
    document.documentElement.lang = currentLanguage;
    
    // Re-render results if they exist
    const resultsDiv = document.getElementById('results');
    if (resultsDiv && resultsDiv.style.display !== 'none') {
        // Store current results and re-render them
        if (window.lastResults) {
            displayResults(window.lastResults);
        }
    }
}