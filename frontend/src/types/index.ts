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
