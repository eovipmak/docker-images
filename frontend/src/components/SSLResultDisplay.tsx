import React from 'react';
import {
  Box,
  Typography,
  Paper,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableRow,
  Alert,
  AlertTitle,
  Grid,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useLanguage } from '../hooks/useLanguage';
import type { SSLCheckResponse } from '../types';

interface SSLResultDisplayProps {
  result: SSLCheckResponse | null;
}

const SSLResultDisplay: React.FC<SSLResultDisplayProps> = ({ result }) => {
  const { t } = useLanguage();

  if (!result) {
    return null;
  }

  if (result.status === 'error') {
    return (
      <Paper elevation={2} sx={{ p: 3, mt: 3 }}>
        <Alert severity="error" icon={<ErrorIcon />}>
          <AlertTitle>{t('statusError')}</AlertTitle>
          {result.error || t('errorOccurred')}
        </Alert>
      </Paper>
    );
  }

  const data = result.data;
  if (!data) {
    return null;
  }

  const ssl = data.ssl || {};
  const ipInfo = data.ip_info || {};

  const getStatusIcon = () => {
    switch (data.sslStatus) {
      case 'success':
        return <CheckCircleIcon color="success" />;
      case 'warning':
        return <WarningIcon color="warning" />;
      default:
        return <ErrorIcon color="error" />;
    }
  };

  const getStatusColor = (): 'success' | 'warning' | 'error' => {
    switch (data.sslStatus) {
      case 'success':
        return 'success';
      case 'warning':
        return 'warning';
      default:
        return 'error';
    }
  };

  return (
    <Paper elevation={2} sx={{ p: 3, mt: 3 }} role="region" aria-label={t('resultsTitle')}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Box display="flex" alignItems="center" gap={1}>
          {getStatusIcon()}
          <Typography variant="h5" component="h2">
            {data.domain || data.ip}
          </Typography>
        </Box>
        <Chip
          label={t(`status${data.sslStatus.charAt(0).toUpperCase() + data.sslStatus.slice(1)}`)}
          color={getStatusColor()}
        />
      </Box>

      <Typography variant="body2" color="text.secondary" gutterBottom>
        {t('ipAddress')}: {data.ip} | {t('port')}: {data.port}
      </Typography>

      {/* SSL Certificate Information */}
      {ssl && Object.keys(ssl).length > 0 && (
        <Box mt={3}>
          <Typography variant="h6" gutterBottom>
            {t('sslCertificate')}
          </Typography>
          <TableContainer>
            <Table size="small" aria-label={t('sslCertificate')}>
              <TableBody>
                {ssl.subject?.commonName && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('subjectCN')}
                    </TableCell>
                    <TableCell>{ssl.subject.commonName}</TableCell>
                  </TableRow>
                )}
                {ssl.subject?.organizationName && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('subjectOrganization')}
                    </TableCell>
                    <TableCell>{ssl.subject.organizationName}</TableCell>
                  </TableRow>
                )}
                {ssl.issuer?.commonName && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('issuer')}
                    </TableCell>
                    <TableCell>{ssl.issuer.commonName}</TableCell>
                  </TableRow>
                )}
                {ssl.notBefore && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('validFrom')}
                    </TableCell>
                    <TableCell>{ssl.notBefore}</TableCell>
                  </TableRow>
                )}
                {ssl.notAfter && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('validUntil')}
                    </TableCell>
                    <TableCell>{ssl.notAfter}</TableCell>
                  </TableRow>
                )}
                {ssl.daysUntilExpiration !== undefined && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('daysUntilExpiration')}
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={`${ssl.daysUntilExpiration} days`}
                        color={ssl.daysUntilExpiration < 30 ? 'error' : 'success'}
                        size="small"
                      />
                    </TableCell>
                  </TableRow>
                )}
                {ssl.tlsVersion && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('tlsVersion')}
                    </TableCell>
                    <TableCell>{ssl.tlsVersion}</TableCell>
                  </TableRow>
                )}
                {ssl.cipherSuite && (
                  <TableRow>
                    <TableCell component="th" scope="row">
                      {t('cipherSuite')}
                    </TableCell>
                    <TableCell>{ssl.cipherSuite}</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </Box>
      )}

      {/* Security Alerts */}
      {ssl.alerts && ssl.alerts.length > 0 && (
        <Alert severity="warning" sx={{ mt: 2 }}>
          <AlertTitle>{t('securityAlerts')}</AlertTitle>
          <ul style={{ margin: 0, paddingLeft: 20 }}>
            {ssl.alerts.map((alert, index) => (
              <li key={index}>{alert}</li>
            ))}
          </ul>
        </Alert>
      )}

      {/* Recommendations */}
      {data.recommendations && data.recommendations.length > 0 && (
        <Alert severity="info" sx={{ mt: 2 }}>
          <AlertTitle>{t('recommendations')}</AlertTitle>
          <ul style={{ margin: 0, paddingLeft: 20 }}>
            {data.recommendations.map((rec, index) => (
              <li key={index}>{rec}</li>
            ))}
          </ul>
        </Alert>
      )}

      {/* Server Information */}
      <Box mt={3}>
        <Typography variant="h6" gutterBottom>
          {t('serverInformation')}
        </Typography>
        <TableContainer>
          <Table size="small" aria-label={t('serverInformation')}>
            <TableBody>
              <TableRow>
                <TableCell component="th" scope="row">
                  {t('server')}
                </TableCell>
                <TableCell>{data.server || t('unknown')}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell component="th" scope="row">
                  {t('sslStatus')}
                </TableCell>
                <TableCell>{data.sslStatus}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Box>

      {/* IP Geolocation */}
      {ipInfo.query && (
        <Box mt={3}>
          <Typography variant="h6" gutterBottom>
            {t('ipGeolocation')}
          </Typography>
          <Grid container spacing={2}>
            {ipInfo.city && (
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  {t('city')}
                </Typography>
                <Typography variant="body1">{ipInfo.city}</Typography>
              </Grid>
            )}
            {ipInfo.region && (
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  {t('region')}
                </Typography>
                <Typography variant="body1">{ipInfo.region}</Typography>
              </Grid>
            )}
            {ipInfo.country && (
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  {t('country')}
                </Typography>
                <Typography variant="body1">{ipInfo.country}</Typography>
              </Grid>
            )}
            {ipInfo.isp && (
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  {t('isp')}
                </Typography>
                <Typography variant="body1">{ipInfo.isp}</Typography>
              </Grid>
            )}
          </Grid>
        </Box>
      )}

      <Typography variant="caption" color="text.secondary" display="block" mt={2}>
        {t('checkedAt')}: {new Date(data.checkedAt).toLocaleString()}
      </Typography>
    </Paper>
  );
};

export default SSLResultDisplay;
