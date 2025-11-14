export interface SSLSubject {
  commonName?: string;
  organizationName?: string;
  organizationalUnitName?: string;
  countryName?: string;
  stateOrProvinceName?: string;
  localityName?: string;
}

export interface SSLIssuer {
  commonName?: string;
  organizationName?: string;
  countryName?: string;
}

export interface SSLInfo {
  subject?: SSLSubject;
  issuer?: SSLIssuer;
  version?: number;
  serialNumber?: string;
  notBefore?: string;
  notAfter?: string;
  daysUntilExpiration?: number;
  signatureAlgorithm?: string;
  tlsVersion?: string;
  cipherSuite?: string;
  subjectAltNames?: string[] | [string, string][];
  alerts?: string[];
}

export interface IPInfo {
  query?: string;
  continent?: string;
  continentCode?: string;
  country?: string;
  countryCode?: string;
  region?: string;
  regionName?: string;
  city?: string;
  district?: string;
  zip?: string;
  isp?: string;
  org?: string;
  as?: string;
  asname?: string;
  reverse?: string;
  lat?: number;
  lon?: number;
  mobile?: boolean;
  proxy?: boolean;
  hosting?: boolean;
}

export interface SSLCheckData {
  domain?: string;
  ip: string;
  port: number;
  server?: string;
  sslStatus: 'success' | 'error' | 'warning';
  serverStatus?: string;
  ipStatus?: string;
  sslErrorType?: string;
  checkedAt: string;
  ssl?: SSLInfo;
  ip_info?: IPInfo;
  recommendations?: string[];
}

export interface SSLCheckResponse {
  status: 'success' | 'error';
  data?: SSLCheckData;
  error?: string;
}

export interface DomainFormData {
  target: string;
}

export interface AlertConfig {
  id: number;
  user_id: number;
  organization_id?: number;
  enabled: boolean;
  webhook_url?: string;
  alert_30_days: boolean;
  alert_7_days: boolean;
  alert_1_day: boolean;
  alert_ssl_errors: boolean;
  alert_geo_changes: boolean;
  alert_cert_expired: boolean;
  email_notifications: boolean;
  email_address?: string;
  created_at: string;
  updated_at: string;
}

export interface AlertConfigUpdate {
  enabled?: boolean;
  webhook_url?: string;
  alert_30_days?: boolean;
  alert_7_days?: boolean;
  alert_1_day?: boolean;
  alert_ssl_errors?: boolean;
  alert_geo_changes?: boolean;
  alert_cert_expired?: boolean;
  email_notifications?: boolean;
  email_address?: string;
}

export interface Alert {
  id: number;
  user_id: number;
  organization_id?: number;
  domain: string;
  alert_type: 'expiring_soon' | 'expired' | 'ssl_error' | 'geo_change' | 'invalid' | 'error';
  severity: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  is_read: boolean;
  is_resolved: boolean;
  created_at: string;
  resolved_at?: string;
}
