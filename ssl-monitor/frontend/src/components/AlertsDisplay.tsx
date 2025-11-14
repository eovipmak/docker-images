import { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  IconButton,
  Chip,
  Alert as MuiAlert,
  CircularProgress,
  Button,
  Tooltip,
} from '@mui/material';
import {
  Warning,
  Error as ErrorIcon,
  Info,
  CheckCircle,
  Delete,
  Refresh,
  MarkEmailRead,
} from '@mui/icons-material';
import { getAlerts, markAlertRead, markAlertResolved, deleteAlert } from '../services/api';
import type { Alert } from '../types';

interface AlertsDisplayProps {
  unresolvedOnly?: boolean;
  limit?: number;
}

export default function AlertsDisplay({ unresolvedOnly = true, limit = 20 }: AlertsDisplayProps) {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAlerts();
  }, [unresolvedOnly, limit]);

  const loadAlerts = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getAlerts(false, unresolvedOnly, limit);
      setAlerts(data);
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to load alerts');
    } finally {
      setLoading(false);
    }
  };

  const handleMarkRead = async (alertId: number) => {
    try {
      await markAlertRead(alertId);
      setAlerts(alerts.map(a => a.id === alertId ? { ...a, is_read: true } : a));
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to mark alert as read');
    }
  };

  const handleResolve = async (alertId: number) => {
    try {
      await markAlertResolved(alertId);
      setAlerts(alerts.map(a => a.id === alertId ? { ...a, is_resolved: true } : a));
      // Refresh if showing only unresolved
      if (unresolvedOnly) {
        setTimeout(loadAlerts, 500);
      }
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to resolve alert');
    }
  };

  const handleDelete = async (alertId: number) => {
    try {
      await deleteAlert(alertId);
      setAlerts(alerts.filter(a => a.id !== alertId));
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to delete alert');
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'error';
      case 'high': return 'error';
      case 'medium': return 'warning';
      case 'low': return 'info';
      default: return 'default';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
      case 'high':
        return <ErrorIcon color="error" />;
      case 'medium':
        return <Warning color="warning" />;
      case 'low':
        return <Info color="info" />;
      default:
        return <Info />;
    }
  };

  const getAlertTypeLabel = (type: string) => {
    switch (type) {
      case 'expiring_soon': return 'Expiring Soon';
      case 'expired': return 'Expired';
      case 'ssl_error': return 'SSL Error';
      case 'geo_change': return 'Geo Change';
      case 'invalid': return 'Invalid';
      default: return type;
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" p={3}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Paper elevation={2} sx={{ p: 2 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Typography variant="h6">
          {unresolvedOnly ? 'Active Alerts' : 'All Alerts'}
        </Typography>
        <Button
          size="small"
          startIcon={<Refresh />}
          onClick={loadAlerts}
        >
          Refresh
        </Button>
      </Box>

      {error && (
        <MuiAlert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </MuiAlert>
      )}

      {alerts.length === 0 ? (
        <Box textAlign="center" py={4}>
          <CheckCircle sx={{ fontSize: 48, color: 'success.main', mb: 1 }} />
          <Typography variant="body1" color="text.secondary">
            No alerts to display
          </Typography>
        </Box>
      ) : (
        <List>
          {alerts.map((alert) => (
            <ListItem
              key={alert.id}
              sx={{
                border: 1,
                borderColor: 'divider',
                borderRadius: 1,
                mb: 1,
                bgcolor: alert.is_read ? 'background.paper' : 'action.hover',
              }}
              secondaryAction={
                <Box>
                  {!alert.is_read && (
                    <Tooltip title="Mark as read">
                      <IconButton
                        edge="end"
                        aria-label="mark read"
                        onClick={() => handleMarkRead(alert.id)}
                        size="small"
                        sx={{ mr: 1 }}
                      >
                        <MarkEmailRead fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  )}
                  {!alert.is_resolved && (
                    <Tooltip title="Resolve">
                      <IconButton
                        edge="end"
                        aria-label="resolve"
                        onClick={() => handleResolve(alert.id)}
                        size="small"
                        sx={{ mr: 1 }}
                      >
                        <CheckCircle fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  )}
                  <Tooltip title="Delete">
                    <IconButton
                      edge="end"
                      aria-label="delete"
                      onClick={() => handleDelete(alert.id)}
                      size="small"
                    >
                      <Delete fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </Box>
              }
            >
              <ListItemIcon>
                {getSeverityIcon(alert.severity)}
              </ListItemIcon>
              <ListItemText
                primary={
                  <Box display="flex" alignItems="center" gap={1}>
                    <Typography variant="body1" fontWeight={alert.is_read ? 'normal' : 'bold'}>
                      {alert.domain}
                    </Typography>
                    <Chip
                      label={getAlertTypeLabel(alert.alert_type)}
                      size="small"
                      color={getSeverityColor(alert.severity) as any}
                    />
                    {alert.is_resolved && (
                      <Chip label="Resolved" size="small" color="success" variant="outlined" />
                    )}
                  </Box>
                }
                secondary={
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      {alert.message}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {new Date(alert.created_at).toLocaleString()}
                    </Typography>
                  </Box>
                }
              />
            </ListItem>
          ))}
        </List>
      )}
    </Paper>
  );
}
