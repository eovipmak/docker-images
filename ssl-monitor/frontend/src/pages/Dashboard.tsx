import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Container,
  Typography,
  Paper,
  Grid,
  Card,
  CardContent,
  Box,
  Chip,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Snackbar,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableRow,
  useTheme,
  useMediaQuery,
  IconButton,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  LinearProgress,
  Select,
  FormControl,
  InputLabel,
  Fade,
  Zoom,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Domain as DomainIcon,
  AccessTime as AccessTimeIcon,
  Refresh as RefreshIcon,
  Delete as DeleteIcon,
  Settings as SettingsIcon,
  NotificationsOff as NotificationsOffIcon,
  Notifications as NotificationsIcon,
  BugReport as BugReportIcon,
  Security as SecurityIcon,
  Schedule as ScheduleIcon,
  FiberManualRecord as FiberManualRecordIcon,
} from '@mui/icons-material';
import { getStats, getHistory, deleteDomain, updateMonitor, testDomainAlert } from '../services/api';
import AlertsDisplay from '../components/AlertsDisplay';

interface Stats {
  total_checks: number;
  successful_checks: number;
  error_checks: number;
  unique_domains: number;
}

interface HistoryItem {
  id: number;
  domain: string;
  status: string;
  checked_at: string;
  ssl_status: string;
  data?: {
    ssl?: {
      daysUntilExpiration?: number;
    };
  };
}

interface SSLInfo {
  daysUntilExpiration?: number;
  notAfter?: string;
  notBefore?: string;
  issuer?: {
    commonName?: string;
    organizationName?: string;
    countryName?: string;
  };
  subject?: {
    commonName?: string;
    organizationName?: string;
  };
  serialNumber?: string;
  signatureAlgorithm?: string;
  tlsVersion?: string;
  cipherSuite?: string;
}

interface DomainStatus {
  domain: string;
  ip?: string;
  port: number;
  status: string;
  ssl_status: string;
  last_checked: string;
  ssl_info: SSLInfo;
  monitor?: {
    alerts_enabled: boolean;
    check_interval: number;
    webhook_url?: string;
  };
  uptime?: {
    uptime_percentage: number | null;
    total_checks: number;
    successful_checks: number;
    failed_checks: number;
    days_tracked: number;
  };
}

const Dashboard: React.FC = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  
  const [stats, setStats] = useState<Stats | null>(null);
  const [recentChecks, setRecentChecks] = useState<HistoryItem[]>([]);
  const [domains, setDomains] = useState<DomainStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDomain, setSelectedDomain] = useState<DomainStatus | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error' | 'warning'>('success');
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const [menuAnchor, setMenuAnchor] = useState<{ element: HTMLElement; domain: DomainStatus } | null>(null);
  const [settingsDialogOpen, setSettingsDialogOpen] = useState(false);
  const [settingsDomain, setSettingsDomain] = useState<DomainStatus | null>(null);
  const [settingsCheckInterval, setSettingsCheckInterval] = useState<number>(3600);
  
  const wsRef = useRef<WebSocket | null>(null);
  const autoRefreshRef = useRef<number | null>(null);

  // Get color based on days until expiration
  const getStatusColor = (daysUntilExpiration?: number): 'success' | 'warning' | 'error' => {
    if (daysUntilExpiration === undefined || daysUntilExpiration === null) {
      return 'error';
    }
    if (daysUntilExpiration > 30) {
      return 'success';
    } else if (daysUntilExpiration >= 7) {
      return 'warning';
    } else {
      return 'error';
    }
  };

  // Format countdown display
  const formatCountdown = (daysUntilExpiration?: number): string => {
    if (daysUntilExpiration === undefined || daysUntilExpiration === null) {
      return 'Unknown';
    }
    if (daysUntilExpiration < 0) {
      return 'Expired';
    }
    if (daysUntilExpiration === 0) {
      return 'Expires today';
    }
    if (daysUntilExpiration === 1) {
      return '1 day';
    }
    return `${daysUntilExpiration} days`;
  };

  // Fetch domains with status
  const fetchDomainsStatus = useCallback(async () => {
    try {
      const token = localStorage.getItem('auth_token');
      if (!token) {
        return;
      }

      const response = await fetch('/api/domains/status', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        if (data.domains) {
          setDomains(data.domains);
          setLastUpdate(new Date());
        }
      }
    } catch (err) {
      console.error('Failed to fetch domain status:', err);
    }
  }, []);

  // Fetch data function
  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [statsResponse, historyResponse] = await Promise.all([
        getStats(),
        getHistory(undefined, 5)
      ]);
      
      setStats(statsResponse.stats);
      const history = historyResponse.history;
      setRecentChecks(Array.isArray(history) ? history : []);
      
      await fetchDomainsStatus();
      
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to fetch data';
      setError(errorMsg);
      showSnackbar(errorMsg, 'error');
    } finally {
      setLoading(false);
    }
  }, [fetchDomainsStatus]);

  // Show snackbar notification
  const showSnackbar = (message: string, severity: 'success' | 'error' | 'warning' = 'success') => {
    setSnackbarMessage(message);
    setSnackbarSeverity(severity);
    setSnackbarOpen(true);
  };

  // Setup WebSocket connection
  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/domains`;
    
    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
        ws.send(JSON.stringify({ type: 'ping' }));
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'update') {
            fetchDomainsStatus();
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
      }
    } catch (err) {
      console.error('Failed to create WebSocket:', err);
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [fetchDomainsStatus]);

  // Setup auto-refresh
  useEffect(() => {
    fetchData();

    autoRefreshRef.current = window.setInterval(() => {
      fetchDomainsStatus();
    }, 30000);

    return () => {
      if (autoRefreshRef.current) {
        window.clearInterval(autoRefreshRef.current);
      }
    };
  }, [fetchData, fetchDomainsStatus]);

  // Handle domain card click
  const handleDomainClick = (domain: DomainStatus) => {
    setSelectedDomain(domain);
  };

  // Handle modal close
  const handleCloseModal = () => {
    setSelectedDomain(null);
  };

  // Handle alert toggle
  const handleToggleAlerts = async (domainName: string, currentStatus: boolean) => {
    try {
      await updateMonitor(domainName, { alerts_enabled: !currentStatus });
      showSnackbar(
        `Alerts ${!currentStatus ? 'enabled' : 'disabled'} for ${domainName}`,
        'success'
      );
      await fetchDomainsStatus();
      setMenuAnchor(null);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update alert settings';
      showSnackbar(errorMsg, 'error');
    }
  };

  // Handle menu open
  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, domain: DomainStatus) => {
    event.stopPropagation();
    setMenuAnchor({ element: event.currentTarget, domain });
  };

  // Handle menu close
  const handleMenuClose = () => {
    setMenuAnchor(null);
  };

  // Handle delete from menu
  const handleDeleteFromMenu = async () => {
    if (menuAnchor) {
      await handleDeleteDomain(menuAnchor.domain.domain);
      setMenuAnchor(null);
    }
  };

  // Handle toggle alerts from menu
  const handleToggleAlertsFromMenu = async () => {
    if (menuAnchor) {
      const alertsEnabled = menuAnchor.domain.monitor?.alerts_enabled ?? true;
      await handleToggleAlerts(menuAnchor.domain.domain, alertsEnabled);
    }
  };

  // Handle test alert from menu
  const handleTestAlertFromMenu = async () => {
    if (!menuAnchor) return;
    
    const domainName = menuAnchor.domain.domain;
    setMenuAnchor(null);
    
    try {
      const result = await testDomainAlert(domainName);
      showSnackbar(result.message || `Test alert sent for ${domainName}`, 'success');
    } catch (err: unknown) {
      const errorMsg = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to send test alert';
      showSnackbar(errorMsg, 'error');
    }
  };

  // Handle settings from menu
  const handleSettingsFromMenu = () => {
    if (!menuAnchor) return;
    
    setSettingsDomain(menuAnchor.domain);
    setSettingsCheckInterval(menuAnchor.domain.monitor?.check_interval || 3600);
    setSettingsDialogOpen(true);
    setMenuAnchor(null);
  };

  // Handle close settings dialog
  const handleCloseSettingsDialog = () => {
    setSettingsDialogOpen(false);
    setSettingsDomain(null);
  };

  // Handle save settings
  const handleSaveSettings = async () => {
    if (!settingsDomain) return;
    
    try {
      await updateMonitor(settingsDomain.domain, { check_interval: settingsCheckInterval });
      showSnackbar(`Settings updated for ${settingsDomain.domain}`, 'success');
      setSettingsDialogOpen(false);
      setSettingsDomain(null);
      await fetchData();
    } catch (err: unknown) {
      const errorMsg = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to update settings';
      showSnackbar(errorMsg, 'error');
    }
  };

  // Handle domain delete
  const handleDeleteDomain = async (domainName: string) => {
    if (!window.confirm(`Are you sure you want to delete monitoring for ${domainName}?`)) {
      return;
    }

    try {
      await deleteDomain(domainName);
      showSnackbar(`Domain ${domainName} deleted successfully`, 'success');
      await fetchData();
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete domain';
      showSnackbar(errorMsg, 'error');
    }
  };

  if (loading) {
    return (
      <Box
        display="flex"
        justifyContent="center"
        alignItems="center"
        minHeight="60vh"
        sx={{
          background: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)',
        }}
      >
        <Box textAlign="center">
          <CircularProgress size={64} sx={{ color: '#6366f1', mb: 2 }} />
          <Typography variant="h6" color="text.secondary">
            Loading dashboard...
          </Typography>
        </Box>
      </Box>
    );
  }

  if (error) {
    return (
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
        <Alert severity="error" sx={{ borderRadius: 3 }}>{error}</Alert>
      </Container>
    );
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)',
        pb: 4,
      }}
    >
      <Container maxWidth="xl" sx={{ pt: 4 }}>
        {/* Header Section */}
        <Fade in={true} timeout={500}>
          <Box 
            display="flex" 
            justifyContent="space-between" 
            alignItems="center" 
            mb={4}
            sx={{
              flexDirection: { xs: 'column', sm: 'row' },
              gap: 2,
            }}
          >
            <Box>
              <Typography 
                variant="h3" 
                component="h1" 
                fontWeight="bold" 
                gutterBottom
                sx={{
                  background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                }}
              >
                SSL Monitor Dashboard
              </Typography>
              <Typography variant="body1" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <FiberManualRecordIcon sx={{ fontSize: 8, color: '#10b981' }} className="animate-pulse" />
                Live monitoring • Last updated {lastUpdate.toLocaleTimeString()}
              </Typography>
            </Box>
            <Button
              startIcon={<RefreshIcon />}
              onClick={fetchData}
              variant="contained"
              size="large"
              sx={{
                borderRadius: 2,
                px: 3,
                py: 1.5,
              }}
            >
              Refresh
            </Button>
          </Box>
        </Fade>

        {/* Stats Cards */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Zoom in={true} timeout={300} style={{ transitionDelay: '0ms' }}>
              <Card 
                sx={{ 
                  height: '100%',
                  background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                  color: 'white',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  '&:hover': {
                    transform: 'translateY(-8px)',
                    boxShadow: '0 20px 25px -5px rgba(99, 102, 241, 0.3), 0 10px 10px -5px rgba(99, 102, 241, 0.2)',
                  }
                }}
              >
                <CardContent sx={{ p: 3 }}>
                  <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                    <SecurityIcon sx={{ fontSize: 40, opacity: 0.9 }} />
                    <Typography variant="h4" fontWeight="bold">
                      {stats?.total_checks || 0}
                    </Typography>
                  </Box>
                  <Typography variant="body2" sx={{ opacity: 0.9, fontWeight: 500 }}>
                    Total SSL Checks
                  </Typography>
                </CardContent>
              </Card>
            </Zoom>
          </Grid>
          
          <Grid item xs={12} sm={6} md={3}>
            <Zoom in={true} timeout={300} style={{ transitionDelay: '100ms' }}>
              <Card 
                sx={{ 
                  height: '100%',
                  background: 'linear-gradient(135deg, #10b981 0%, #34d399 100%)',
                  color: 'white',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  '&:hover': {
                    transform: 'translateY(-8px)',
                    boxShadow: '0 20px 25px -5px rgba(16, 185, 129, 0.3), 0 10px 10px -5px rgba(16, 185, 129, 0.2)',
                  }
                }}
              >
                <CardContent sx={{ p: 3 }}>
                  <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                    <CheckCircleIcon sx={{ fontSize: 40, opacity: 0.9 }} />
                    <Typography variant="h4" fontWeight="bold">
                      {stats?.successful_checks || 0}
                    </Typography>
                  </Box>
                  <Typography variant="body2" sx={{ opacity: 0.9, fontWeight: 500 }}>
                    Valid Certificates
                  </Typography>
                </CardContent>
              </Card>
            </Zoom>
          </Grid>
          
          <Grid item xs={12} sm={6} md={3}>
            <Zoom in={true} timeout={300} style={{ transitionDelay: '200ms' }}>
              <Card 
                sx={{ 
                  height: '100%',
                  background: 'linear-gradient(135deg, #ef4444 0%, #f87171 100%)',
                  color: 'white',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  '&:hover': {
                    transform: 'translateY(-8px)',
                    boxShadow: '0 20px 25px -5px rgba(239, 68, 68, 0.3), 0 10px 10px -5px rgba(239, 68, 68, 0.2)',
                  }
                }}
              >
                <CardContent sx={{ p: 3 }}>
                  <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                    <ErrorIcon sx={{ fontSize: 40, opacity: 0.9 }} />
                    <Typography variant="h4" fontWeight="bold">
                      {stats?.error_checks || 0}
                    </Typography>
                  </Box>
                  <Typography variant="body2" sx={{ opacity: 0.9, fontWeight: 500 }}>
                    Certificate Errors
                  </Typography>
                </CardContent>
              </Card>
            </Zoom>
          </Grid>
          
          <Grid item xs={12} sm={6} md={3}>
            <Zoom in={true} timeout={300} style={{ transitionDelay: '300ms' }}>
              <Card 
                sx={{ 
                  height: '100%',
                  background: 'linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%)',
                  color: 'white',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  '&:hover': {
                    transform: 'translateY(-8px)',
                    boxShadow: '0 20px 25px -5px rgba(59, 130, 246, 0.3), 0 10px 10px -5px rgba(59, 130, 246, 0.2)',
                  }
                }}
              >
                <CardContent sx={{ p: 3 }}>
                  <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                    <DomainIcon sx={{ fontSize: 40, opacity: 0.9 }} />
                    <Typography variant="h4" fontWeight="bold">
                      {stats?.unique_domains || 0}
                    </Typography>
                  </Box>
                  <Typography variant="body2" sx={{ opacity: 0.9, fontWeight: 500 }}>
                    Monitored Domains
                  </Typography>
                </CardContent>
              </Card>
            </Zoom>
          </Grid>
        </Grid>

        {/* Monitored Domains Section */}
        <Fade in={true} timeout={600}>
          <Box mb={3}>
            <Typography variant="h5" gutterBottom fontWeight="bold" sx={{ mb: 1 }}>
              Monitored Domains
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Real-time SSL certificate monitoring for your domains
            </Typography>
          </Box>
        </Fade>
        
        {domains.length === 0 ? (
          <Alert 
            severity="info" 
            sx={{ 
              mb: 4, 
              borderRadius: 3,
              '& .MuiAlert-icon': {
                fontSize: 28,
              }
            }}
          >
            No domains are being monitored. Add a domain to get started.
          </Alert>
        ) : (
          <Grid container spacing={3} sx={{ mb: 4 }}>
            {domains.map((domain, index) => {
              const statusColor = getStatusColor(domain.ssl_info?.daysUntilExpiration);
              const countdown = formatCountdown(domain.ssl_info?.daysUntilExpiration);
              
              return (
                <Grid item xs={12} sm={6} md={4} lg={3} key={`${domain.domain}-${index}`}>
                  <Fade in={true} timeout={400} style={{ transitionDelay: `${index * 50}ms` }}>
                    <Card 
                      sx={{ 
                        height: '100%',
                        borderLeft: `4px solid ${
                          statusColor === 'success' 
                            ? '#10b981'
                            : statusColor === 'warning' 
                            ? '#f59e0b'
                            : '#ef4444'
                        }`,
                        transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                        cursor: 'pointer',
                        '&:hover': {
                          transform: 'translateY(-8px)',
                          boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
                        }
                      }}
                      onClick={() => handleDomainClick(domain)}
                    >
                      <CardContent sx={{ p: 3 }}>
                        <Box display="flex" alignItems="flex-start" justifyContent="space-between" mb={2}>
                          <Box display="flex" alignItems="center" gap={1.5} flex={1} minWidth={0}>
                            <Box
                              sx={{
                                p: 1,
                                borderRadius: 2,
                                bgcolor: statusColor === 'success' 
                                  ? 'rgba(16, 185, 129, 0.1)'
                                  : statusColor === 'warning'
                                  ? 'rgba(245, 158, 11, 0.1)'
                                  : 'rgba(239, 68, 68, 0.1)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                              }}
                            >
                              <DomainIcon 
                                sx={{ 
                                  fontSize: 24,
                                  color: statusColor === 'success' 
                                    ? '#10b981'
                                    : statusColor === 'warning'
                                    ? '#f59e0b'
                                    : '#ef4444',
                                }} 
                              />
                            </Box>
                            <Box flex={1} minWidth={0}>
                              <Typography 
                                variant="h6" 
                                component="div"
                                sx={{
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  whiteSpace: 'nowrap',
                                  fontWeight: 600,
                                }}
                              >
                                {domain.domain}
                              </Typography>
                              <Typography variant="caption" color="text.secondary">
                                Port {domain.port}
                              </Typography>
                            </Box>
                          </Box>
                          <IconButton
                            size="small"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleMenuOpen(e, domain);
                            }}
                            sx={{ 
                              ml: 1,
                              '&:hover': {
                                bgcolor: 'action.hover',
                              }
                            }}
                          >
                            <SettingsIcon fontSize="small" />
                          </IconButton>
                        </Box>
                        
                        <Box display="flex" alignItems="center" gap={1} mb={2}>
                          <AccessTimeIcon fontSize="small" color="action" />
                          <Typography variant="body2" color="text.secondary" fontWeight={500}>
                            {countdown}
                          </Typography>
                        </Box>
                        
                        <Box display="flex" gap={1} flexWrap="wrap" mb={2}>
                          <Chip
                            label={domain.ssl_status === 'success' ? 'Valid SSL' : 'SSL Error'}
                            color={domain.ssl_status === 'success' ? 'success' : 'error'}
                            size="small"
                            sx={{ fontWeight: 500 }}
                          />
                          {domain.ssl_info?.daysUntilExpiration !== undefined && (
                            <Chip
                              label={`${domain.ssl_info.daysUntilExpiration}d left`}
                              color={statusColor}
                              size="small"
                              sx={{ fontWeight: 500 }}
                            />
                          )}
                        </Box>
                        
                        {domain.uptime && domain.uptime.uptime_percentage !== null && (
                          <Box>
                            <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                              <Typography variant="caption" color="text.secondary" fontWeight={600}>
                                30-Day Uptime
                              </Typography>
                              <Typography 
                                variant="caption" 
                                fontWeight="bold"
                                sx={{
                                  color: domain.uptime.uptime_percentage >= 99 
                                    ? '#10b981'
                                    : domain.uptime.uptime_percentage >= 95 
                                    ? '#f59e0b'
                                    : '#ef4444'
                                }}
                              >
                                {domain.uptime.uptime_percentage.toFixed(2)}%
                              </Typography>
                            </Box>
                            <LinearProgress 
                              variant="determinate" 
                              value={domain.uptime.uptime_percentage} 
                              sx={{
                                height: 8,
                                borderRadius: 4,
                                bgcolor: 'rgba(0, 0, 0, 0.05)',
                                '& .MuiLinearProgress-bar': {
                                  borderRadius: 4,
                                  bgcolor: domain.uptime.uptime_percentage >= 99 
                                    ? '#10b981'
                                    : domain.uptime.uptime_percentage >= 95 
                                    ? '#f59e0b'
                                    : '#ef4444'
                                }
                              }}
                            />
                          </Box>
                        )}
                      </CardContent>
                    </Card>
                  </Fade>
                </Grid>
              );
            })}
          </Grid>
        )}

        {/* Recent Checks and Alerts */}
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Fade in={true} timeout={800}>
              <AlertsDisplay unresolvedOnly={true} limit={10} />
            </Fade>
          </Grid>
          <Grid item xs={12} md={6}>
            <Fade in={true} timeout={800}>
              <Paper 
                elevation={0}
                sx={{ 
                  p: 3,
                  borderRadius: 3,
                  background: 'white',
                  border: '1px solid',
                  borderColor: 'divider',
                }}
              >
                <Box display="flex" alignItems="center" gap={1.5} mb={3}>
                  <ScheduleIcon color="primary" />
                  <Typography variant="h6" fontWeight="bold">
                    Recent SSL Checks
                  </Typography>
                </Box>
                {recentChecks.length === 0 ? (
                  <Box textAlign="center" py={4}>
                    <CheckCircleIcon sx={{ fontSize: 48, color: 'success.main', mb: 1, opacity: 0.5 }} />
                    <Typography variant="body2" color="text.secondary">
                      No checks found. Start by adding a domain to monitor.
                    </Typography>
                  </Box>
                ) : (
                  <Box display="flex" flexDirection="column" gap={2}>
                    {recentChecks.map((check) => (
                      <Card 
                        key={check.id}
                        variant="outlined"
                        sx={{
                          borderRadius: 2,
                          transition: 'all 0.2s',
                          '&:hover': {
                            boxShadow: 2,
                            transform: 'translateX(4px)',
                          }
                        }}
                      >
                        <CardContent sx={{ p: 2 }}>
                          <Box display="flex" justifyContent="space-between" alignItems="center" flexWrap="wrap" gap={1}>
                            <Box display="flex" alignItems="center" gap={2}>
                              {check.status === 'success' ? (
                                <CheckCircleIcon color="success" />
                              ) : (
                                <ErrorIcon color="error" />
                              )}
                              <Box>
                                <Typography variant="body1" fontWeight={600}>
                                  {check.domain || 'N/A'}
                                </Typography>
                                <Typography variant="caption" color="text.secondary">
                                  {new Date(check.checked_at).toLocaleString()}
                                </Typography>
                              </Box>
                            </Box>
                            <Chip
                              label={check.status === 'success' ? 'Success' : 'Error'}
                              color={check.status === 'success' ? 'success' : 'error'}
                              size="small"
                              sx={{ fontWeight: 500 }}
                            />
                          </Box>
                        </CardContent>
                      </Card>
                    ))}
                  </Box>
                )}
              </Paper>
            </Fade>
          </Grid>
        </Grid>

        {/* Certificate Details Modal */}
        <Dialog 
          open={!!selectedDomain} 
          onClose={handleCloseModal}
          maxWidth="md"
          fullWidth
          fullScreen={isMobile}
          PaperProps={{
            sx: {
              borderRadius: 3,
            }
          }}
        >
          <DialogTitle>
            <Box display="flex" alignItems="center" gap={2}>
              <Box
                sx={{
                  p: 1.5,
                  borderRadius: 2,
                  bgcolor: 'primary.main',
                  color: 'white',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <SecurityIcon />
              </Box>
              <Box>
                <Typography variant="h6" fontWeight="bold">
                  SSL Certificate Details
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {selectedDomain?.domain}
                </Typography>
              </Box>
            </Box>
          </DialogTitle>
          <DialogContent>
            {selectedDomain && (
              <TableContainer>
                <Table>
                  <TableBody>
                    <TableRow>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                        Certificate Status
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={selectedDomain.ssl_status === 'success' ? 'Valid' : 'Error'}
                          color={selectedDomain.ssl_status === 'success' ? 'success' : 'error'}
                          size="small"
                        />
                      </TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                        Domain
                      </TableCell>
                      <TableCell>{selectedDomain.domain}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                        IP Address
                      </TableCell>
                      <TableCell>{selectedDomain.ip || 'N/A'}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                        Port
                      </TableCell>
                      <TableCell>{selectedDomain.port}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                        Days Until Expiration
                      </TableCell>
                      <TableCell>
                        <Box display="flex" alignItems="center" gap={1}>
                          <Typography>
                            {formatCountdown(selectedDomain.ssl_info?.daysUntilExpiration)}
                          </Typography>
                          {selectedDomain.ssl_info?.daysUntilExpiration !== undefined && (
                            <Chip
                              label={`${selectedDomain.ssl_info.daysUntilExpiration}d`}
                              color={getStatusColor(selectedDomain.ssl_info.daysUntilExpiration)}
                              size="small"
                            />
                          )}
                        </Box>
                      </TableCell>
                    </TableRow>
                    {selectedDomain.ssl_info?.notAfter && (
                      <TableRow>
                        <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                          Expires On
                        </TableCell>
                        <TableCell>
                          {new Date(selectedDomain.ssl_info.notAfter).toLocaleString()}
                        </TableCell>
                      </TableRow>
                    )}
                    {selectedDomain.ssl_info?.issuer?.organizationName && (
                      <TableRow>
                        <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                          Issuer
                        </TableCell>
                        <TableCell>
                          {selectedDomain.ssl_info.issuer.organizationName || selectedDomain.ssl_info.issuer.commonName}
                        </TableCell>
                      </TableRow>
                    )}
                    <TableRow>
                      <TableCell component="th" scope="row" sx={{ fontWeight: 600 }}>
                        Last Checked
                      </TableCell>
                      <TableCell>
                        {new Date(selectedDomain.last_checked).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={handleCloseModal} variant="outlined">
              Close
            </Button>
          </DialogActions>
        </Dialog>

        {/* Monitor Settings Dialog */}
        <Dialog
          open={settingsDialogOpen}
          onClose={handleCloseSettingsDialog}
          maxWidth="sm"
          fullWidth
          disableEnforceFocus
          disableAutoFocus
          PaperProps={{
            sx: {
              borderRadius: 3,
            }
          }}
        >
          <DialogTitle>
            <Typography variant="h6" fontWeight="bold">
              Monitor Settings
            </Typography>
            {settingsDomain && (
              <Typography variant="body2" color="text.secondary">
                {settingsDomain.domain}
              </Typography>
            )}
          </DialogTitle>
          <DialogContent>
            <Box sx={{ pt: 2 }}>
              <FormControl fullWidth>
                <InputLabel id="check-interval-label">Check Interval</InputLabel>
                <Select
                  labelId="check-interval-label"
                  value={settingsCheckInterval}
                  label="Check Interval"
                  onChange={(e) => setSettingsCheckInterval(Number(e.target.value))}
                  MenuProps={{
                    disableScrollLock: true,
                    disablePortal: true,
                    PaperProps: {
                      sx: {
                        maxHeight: 300,
                      }
                    },
                    anchorOrigin: {
                      vertical: 'bottom',
                      horizontal: 'left',
                    },
                    transformOrigin: {
                      vertical: 'top',
                      horizontal: 'left',
                    },
                  }}
                >
                  <MenuItem value={3600}>1 hour</MenuItem>
                  <MenuItem value={10800}>3 hours</MenuItem>
                  <MenuItem value={43200}>12 hours</MenuItem>
                  <MenuItem value={86400}>24 hours (1 day)</MenuItem>
                </Select>
              </FormControl>
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                How often to check the SSL certificate status
              </Typography>
            </Box>
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={handleCloseSettingsDialog} variant="outlined">
              Cancel
            </Button>
            <Button onClick={handleSaveSettings} variant="contained">
              Save
            </Button>
          </DialogActions>
        </Dialog>

        {/* Snackbar for notifications */}
        <Snackbar
          open={snackbarOpen}
          autoHideDuration={6000}
          onClose={() => setSnackbarOpen(false)}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        >
          <Alert 
            onClose={() => setSnackbarOpen(false)} 
            severity={snackbarSeverity}
            sx={{ 
              width: '100%',
              borderRadius: 2,
            }}
          >
            {snackbarMessage}
          </Alert>
        </Snackbar>

        {/* Domain Actions Menu */}
        <Menu
          anchorEl={menuAnchor?.element}
          open={Boolean(menuAnchor)}
          onClose={handleMenuClose}
          PaperProps={{
            sx: {
              borderRadius: 2,
              mt: 1,
            }
          }}
        >
          <MenuItem onClick={handleToggleAlertsFromMenu}>
            <ListItemIcon>
              {menuAnchor?.domain.monitor?.alerts_enabled === false ? (
                <NotificationsIcon fontSize="small" />
              ) : (
                <NotificationsOffIcon fontSize="small" />
              )}
            </ListItemIcon>
            <ListItemText>
              {menuAnchor?.domain.monitor?.alerts_enabled === false
                ? 'Enable Alerts'
                : 'Disable Alerts'}
            </ListItemText>
          </MenuItem>
          <MenuItem onClick={handleSettingsFromMenu}>
            <ListItemIcon>
              <SettingsIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>Monitor Settings</ListItemText>
          </MenuItem>
          <MenuItem onClick={handleTestAlertFromMenu}>
            <ListItemIcon>
              <BugReportIcon fontSize="small" color="primary" />
            </ListItemIcon>
            <ListItemText>Send Test Alert</ListItemText>
          </MenuItem>
          <MenuItem onClick={handleDeleteFromMenu}>
            <ListItemIcon>
              <DeleteIcon fontSize="small" color="error" />
            </ListItemIcon>
            <ListItemText>Delete Domain</ListItemText>
          </MenuItem>
        </Menu>
      </Container>
    </Box>
  );
};

export default Dashboard;
