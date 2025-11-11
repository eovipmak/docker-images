// Translation data for the SSL Checker application
const translations = {
    vi: {
        // Header
        title: "🔒 Công cụ Kiểm tra Chứng chỉ SSL",
        subtitle: "Kiểm tra thông tin chứng chỉ SSL, chi tiết máy chủ và vị trí địa lý IP",
        
        // Form
        formTitle: "Kiểm tra Chứng chỉ SSL",
        targetLabel: "Tên miền hoặc Địa chỉ IP",
        targetPlaceholder: "example.com:443, 93.184.216.34:8443, hoặc [::1]:443",
        helpText: "Nhập tên miền hoặc địa chỉ IP (IPv4/IPv6). Cổng tùy chọn theo định dạng domain[:port], IP[:port], hoặc [IPv6]:port (mặc định: 443)",
        checkButton: "Kiểm tra Chứng chỉ",
        
        // Results
        resultsTitle: "Kết quả",
        result: "Kết quả",
        
        // Status badges
        statusSuccess: "Thành công",
        statusError: "Lỗi",
        statusWarning: "Cảnh báo",
        
        // Sections
        sslCertificate: "🔒 Chứng chỉ SSL",
        serverInformation: "🖥️ Thông tin Máy chủ",
        ipGeolocation: "🌍 Vị trí Địa lý IP",
        securityAlerts: "⚠️ Cảnh báo Bảo mật",
        recommendations: "💡 Khuyến nghị",
        
        // SSL Certificate Fields
        subjectCN: "Subject CN (Tên chung)",
        subjectOrganization: "Tổ chức Subject",
        subjectOrgUnit: "Đơn vị Tổ chức Subject",
        subjectCountry: "Quốc gia Subject",
        subjectState: "Tỉnh/Bang Subject",
        subjectLocality: "Địa phương Subject",
        issuer: "Nhà phát hành",
        issuerOrg: "Tổ chức Nhà phát hành",
        issuerCountry: "Quốc gia Nhà phát hành",
        version: "Phiên bản",
        serialNumber: "Số Serial",
        validFrom: "Có hiệu lực từ",
        validUntil: "Có hiệu lực đến",
        daysUntilExpiration: "Số ngày đến khi hết hạn",
        tlsVersion: "Phiên bản TLS",
        cipherSuite: "Bộ mã hóa",
        signatureAlgorithm: "Thuật toán Chữ ký",
        subjectAltNames: "Tên thay thế Subject (SAN)",
        
        // Server Information Fields
        ipAddress: "Địa chỉ IP",
        port: "Cổng",
        server: "Máy chủ",
        
        // IP Geolocation Fields
        continent: "Châu lục",
        continentCode: "Mã châu lục",
        country: "Quốc gia",
        countryCode: "Mã quốc gia",
        region: "Vùng",
        regionName: "Tên vùng",
        city: "Thành phố",
        district: "Quận/Huyện",
        zip: "Mã bưu điện",
        coordinates: "Tọa độ (Vĩ độ, Kinh độ)",
        isp: "Nhà cung cấp dịch vụ Internet",
        org: "Tổ chức",
        asn: "Số hệ thống tự trị (AS)",
        asname: "Tên AS",
        reverse: "DNS ngược",
        mobile: "Di động",
        proxy: "Proxy",
        hosting: "Lưu trữ",
        
        // Status Fields
        sslStatus: "Trạng thái SSL",
        serverStatus: "Trạng thái Máy chủ",
        ipStatus: "Trạng thái IP",
        errorType: "Loại lỗi",
        checkedAt: "Kiểm tra lúc",
        
        // Common
        yes: "Có",
        no: "Không",
        unknown: "Không rõ",
        notAvailable: "N/A",
        
        // Errors
        errorOccurred: "Đã xảy ra lỗi",
        provideDomain: "Vui lòng cung cấp tên miền hoặc địa chỉ IP",
        invalidTarget: "Vui lòng cung cấp tên miền hoặc địa chỉ IP hợp lệ",
        checkFailed: "Không thể kiểm tra chứng chỉ SSL. Vui lòng thử lại.",
        
        // Footer
        footerText: "🔒 SSL Checker API v2.0.0",
        builtWith: "Được xây dựng với",
        by: "bởi"
    },
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

// Current language (default: Vietnamese)
let currentLanguage = 'vi';

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
    
    // Update language toggle button text
    const langToggle = document.getElementById('languageToggle');
    if (langToggle) {
        langToggle.textContent = currentLanguage === 'vi' ? 'English' : 'Tiếng Việt';
    }
    
    // Re-render results if they exist
    const resultsDiv = document.getElementById('results');
    if (resultsDiv && resultsDiv.style.display !== 'none') {
        // Store current results and re-render them
        if (window.lastResults) {
            displayResults(window.lastResults);
        }
    }
}
