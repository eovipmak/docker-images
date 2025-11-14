type Language = 'en';

export const translations: Record<Language, Record<string, string>> = {
  en: {
    // Header
    title: '🔒 SSL Certificate Checker',
    subtitle: 'Check SSL certificate information, server details, and IP geolocation',
    
    // Navigation
    login: 'Login',
    dashboard: 'Dashboard',
    addDomain: 'Add Domain',
    alerts: 'Alerts',
    alertSettings: 'Alert Settings',
    
    // Form
    formTitle: 'SSL Certificate Check',
    targetLabel: 'Domain or IP Address',
    targetPlaceholder: 'example.com:443, 93.184.216.34:8443, or [::1]:443',
    helpText: 'Enter a domain or IP address (IPv4/IPv6). Optional port in format domain[:port], IP[:port], or [IPv6]:port (default: 443)',
    checkButton: 'Check Certificate',
    
    // Results
    resultsTitle: 'Results',
    result: 'Result',
    
    // Status badges
    statusSuccess: 'Success',
    statusError: 'Error',
    statusWarning: 'Warning',
    
    // Sections
    sslCertificate: '🔒 SSL Certificate',
    serverInformation: '🖥️ Server Information',
    ipGeolocation: '🌍 IP Geolocation',
    securityAlerts: '⚠️ Security Alerts',
    recommendations: '💡 Recommendations',
    
    // SSL Certificate Fields
    subjectCN: 'Subject CN (Common Name)',
    subjectOrganization: 'Subject Organization',
    subjectOrgUnit: 'Subject Organizational Unit',
    subjectCountry: 'Subject Country',
    subjectState: 'Subject State/Province',
    subjectLocality: 'Subject Locality',
    issuer: 'Issuer',
    issuerOrg: 'Issuer Organization',
    issuerCountry: 'Issuer Country',
    version: 'Version',
    serialNumber: 'Serial Number',
    validFrom: 'Valid From',
    validUntil: 'Valid Until',
    daysUntilExpiration: 'Days Until Expiration',
    tlsVersion: 'TLS Version',
    cipherSuite: 'Cipher Suite',
    signatureAlgorithm: 'Signature Algorithm',
    subjectAltNames: 'Subject Alternative Names (SAN)',
    
    // Server Information Fields
    ipAddress: 'IP Address',
    port: 'Port',
    server: 'Server',
    
    // IP Geolocation Fields
    continent: 'Continent',
    continentCode: 'Continent Code',
    country: 'Country',
    countryCode: 'Country Code',
    region: 'Region',
    regionName: 'Region Name',
    city: 'City',
    district: 'District',
    zip: 'Zip Code',
    coordinates: 'Coordinates (Lat, Lon)',
    isp: 'Internet Service Provider',
    org: 'Organization',
    asn: 'Autonomous System Number (AS)',
    asname: 'AS Name',
    reverse: 'Reverse DNS',
    mobile: 'Mobile',
    proxy: 'Proxy',
    hosting: 'Hosting',
    
    // Status Fields
    sslStatus: 'SSL Status',
    serverStatus: 'Server Status',
    ipStatus: 'IP Status',
    errorType: 'Error Type',
    checkedAt: 'Checked At',
    
    // Common
    yes: 'Yes',
    no: 'No',
    unknown: 'Unknown',
    notAvailable: 'N/A',
    loading: 'Loading...',
    
    // Errors
    errorOccurred: 'An error occurred',
    provideDomain: 'Please provide a domain name or IP address',
    invalidTarget: 'Please provide a valid domain name or IP address',
    checkFailed: 'Failed to check SSL certificate. Please try again.',
  },
};
