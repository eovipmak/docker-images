import { useState, useEffect, useRef } from 'react';
import {
  Container,
  Paper,
  Typography,
  Box,
  Switch,
  FormControlLabel,
  TextField,
  Button,
  Alert,
  CircularProgress,
  Divider,
  Grid,
  Chip,
} from '@mui/material';
import { Save as SaveIcon, NotificationsActive, Send as SendIcon } from '@mui/icons-material';
import { getAlertConfig, updateAlertConfig, testWebhook } from '../services/api';
import type { AlertConfig, AlertConfigUpdate } from '../types';

export default function AlertSettings() {
  const [config, setConfig] = useState<AlertConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testingWebhook, setTestingWebhook] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [webhookTestResult, setWebhookTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const successTimerRef = useRef<number | null>(null);
  const webhookTestTimerRef = useRef<number | null>(null);

  useEffect(() => {
    loadConfig();
    
    // Cleanup on unmount
    return () => {
      if (successTimerRef.current !== null) {
        window.clearTimeout(successTimerRef.current);
      }
      if (webhookTestTimerRef.current !== null) {
        window.clearTimeout(webhookTestTimerRef.current);
      }
    };
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getAlertConfig();
      setConfig(data);
    } catch (err: unknown) {
      const message = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to load alert configuration';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!config) return;

    try {
      setSaving(true);
      setError(null);
      setSuccess(false);
      
      // Clear any existing timer
      if (successTimerRef.current !== null) {
        window.clearTimeout(successTimerRef.current);
      }

      const updateData: AlertConfigUpdate = {
        enabled: config.enabled,
        webhook_url: config.webhook_url || undefined,
        alert_30_days: config.alert_30_days,
        alert_7_days: config.alert_7_days,
        alert_1_day: config.alert_1_day,
        alert_ssl_errors: config.alert_ssl_errors,
        alert_geo_changes: config.alert_geo_changes,
        alert_cert_expired: config.alert_cert_expired,
        email_notifications: config.email_notifications,
        email_address: config.email_address || undefined,
      };

      const updatedConfig = await updateAlertConfig(updateData);
      setConfig(updatedConfig);
      setSuccess(true);
      
      // Set timer with cleanup tracking
      successTimerRef.current = window.setTimeout(() => {
        setSuccess(false);
        successTimerRef.current = null;
      }, 3000);
    } catch (err: unknown) {
      const message = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to save configuration';
      setError(message);
    } finally {
      setSaving(false);
    }
  };

  const handleChange = (field: keyof AlertConfig, value: string | boolean) => {
    if (!config) return;
    setConfig({ ...config, [field]: value });
  };

  const handleTestWebhook = async () => {
    if (!config?.webhook_url) {
      setWebhookTestResult({ 
        success: false, 
        message: 'Please enter a webhook URL first' 
      });
      return;
    }

    try {
      setTestingWebhook(true);
      setWebhookTestResult(null);
      
      // Clear any existing timer
      if (webhookTestTimerRef.current !== null) {
        window.clearTimeout(webhookTestTimerRef.current);
      }

      const result = await testWebhook();
      setWebhookTestResult({ 
        success: true, 
        message: result.message || 'Test notification sent successfully!' 
      });
      
      // Set timer with cleanup tracking
      webhookTestTimerRef.current = window.setTimeout(() => {
        setWebhookTestResult(null);
        webhookTestTimerRef.current = null;
      }, 5000);
    } catch (err: unknown) {
      const message = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to send test notification';
      setWebhookTestResult({ 
        success: false, 
        message 
      });
      
      // Set timer with cleanup tracking
      webhookTestTimerRef.current = window.setTimeout(() => {
        setWebhookTestResult(null);
        webhookTestTimerRef.current = null;
      }, 5000);
    } finally {
      setTestingWebhook(false);
    }
  };

  if (loading) {
    return (
      <Container maxWidth="md" sx={{ mt: 4, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Container>
    );
  }

  if (!config) {
    return (
      <Container maxWidth="md" sx={{ mt: 4 }}>
        <Alert severity="error">Failed to load alert configuration</Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ mt: 4, mb: 4 }}>
      <Paper elevation={3} sx={{ p: 4 }}>
        <Box display="flex" alignItems="center" mb={3}>
          <NotificationsActive sx={{ mr: 2, fontSize: 32 }} color="primary" />
          <Typography variant="h4" component="h1">
            Alert Settings
          </Typography>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}

        {success && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Alert configuration saved successfully!
          </Alert>
        )}

        {/* Master Enable/Disable */}
        <Box mb={3}>
          <FormControlLabel
            control={
              <Switch
                checked={config.enabled}
                onChange={(e) => handleChange('enabled', e.target.checked)}
                color="primary"
                size="medium"
              />
            }
            label={
              <Box>
                <Typography variant="h6">Enable Alerts</Typography>
                <Typography variant="body2" color="text.secondary">
                  Master switch to enable or disable all alert notifications
                </Typography>
              </Box>
            }
          />
        </Box>

        <Divider sx={{ my: 3 }} />

        {/* Certificate Expiration Alerts */}
        <Box mb={3}>
          <Typography variant="h6" gutterBottom>
            Certificate Expiration Alerts
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Get notified when SSL certificates are about to expire
          </Typography>

          <Grid container spacing={2}>
            <Grid item xs={12} sm={4}>
              <FormControlLabel
                control={
                  <Switch
                    checked={config.alert_30_days}
                    onChange={(e) => handleChange('alert_30_days', e.target.checked)}
                    disabled={!config.enabled}
                  />
                }
                label={
                  <Box>
                    <Typography variant="body1">30 Days</Typography>
                    <Chip label="Medium" size="small" color="warning" />
                  </Box>
                }
              />
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControlLabel
                control={
                  <Switch
                    checked={config.alert_7_days}
                    onChange={(e) => handleChange('alert_7_days', e.target.checked)}
                    disabled={!config.enabled}
                  />
                }
                label={
                  <Box>
                    <Typography variant="body1">7 Days</Typography>
                    <Chip label="High" size="small" color="error" />
                  </Box>
                }
              />
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControlLabel
                control={
                  <Switch
                    checked={config.alert_1_day}
                    onChange={(e) => handleChange('alert_1_day', e.target.checked)}
                    disabled={!config.enabled}
                  />
                }
                label={
                  <Box>
                    <Typography variant="body1">1 Day</Typography>
                    <Chip label="Critical" size="small" sx={{ bgcolor: '#d32f2f', color: 'white' }} />
                  </Box>
                }
              />
            </Grid>
          </Grid>
        </Box>

        <Divider sx={{ my: 3 }} />

        {/* Other Alert Types */}
        <Box mb={3}>
          <Typography variant="h6" gutterBottom>
            Other Alert Types
          </Typography>

          <Box mb={2}>
            <FormControlLabel
              control={
                <Switch
                  checked={config.alert_cert_expired}
                  onChange={(e) => handleChange('alert_cert_expired', e.target.checked)}
                  disabled={!config.enabled}
                />
              }
              label={
                <Box>
                  <Typography variant="body1">Expired Certificates</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Alert when certificate has already expired
                  </Typography>
                </Box>
              }
            />
          </Box>

          <Box mb={2}>
            <FormControlLabel
              control={
                <Switch
                  checked={config.alert_ssl_errors}
                  onChange={(e) => handleChange('alert_ssl_errors', e.target.checked)}
                  disabled={!config.enabled}
                />
              }
              label={
                <Box>
                  <Typography variant="body1">SSL Errors</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Alert on SSL validation errors or issues
                  </Typography>
                </Box>
              }
            />
          </Box>

          <Box mb={2}>
            <FormControlLabel
              control={
                <Switch
                  checked={config.alert_geo_changes}
                  onChange={(e) => handleChange('alert_geo_changes', e.target.checked)}
                  disabled={!config.enabled}
                />
              }
              label={
                <Box>
                  <Typography variant="body1">Geolocation Changes</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Alert when server geolocation changes (e.g., different country)
                  </Typography>
                </Box>
              }
            />
          </Box>
        </Box>

        <Divider sx={{ my: 3 }} />

        {/* Webhook Configuration */}
        <Box mb={3}>
          <Typography variant="h6" gutterBottom>
            Webhook Notifications
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Send alert notifications to a webhook URL (e.g., Slack, Discord, Teams)
          </Typography>

          <TextField
            fullWidth
            label="Webhook URL"
            placeholder="https://hooks.slack.com/services/..."
            value={config.webhook_url || ''}
            onChange={(e) => handleChange('webhook_url', e.target.value)}
            disabled={!config.enabled}
            helperText="Optional. Leave empty to disable webhook notifications."
            sx={{ mb: 2 }}
          />

          {webhookTestResult && (
            <Alert 
              severity={webhookTestResult.success ? 'success' : 'error'} 
              sx={{ mb: 2 }}
              onClose={() => setWebhookTestResult(null)}
            >
              {webhookTestResult.message}
            </Alert>
          )}

          <Button
            variant="outlined"
            color="primary"
            startIcon={testingWebhook ? <CircularProgress size={20} /> : <SendIcon />}
            onClick={handleTestWebhook}
            disabled={!config.webhook_url || testingWebhook || !config.enabled}
          >
            {testingWebhook ? 'Sending Test...' : 'Test Webhook'}
          </Button>
        </Box>

        {/* Save Button */}
        <Box mt={4} display="flex" justifyContent="flex-end">
          <Button
            variant="contained"
            color="primary"
            size="large"
            startIcon={saving ? <CircularProgress size={20} /> : <SaveIcon />}
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save Configuration'}
          </Button>
        </Box>

        {/* Configuration Info */}
        <Box mt={3} p={2} bgcolor="grey.100" borderRadius={1}>
          <Typography variant="caption" color="text.secondary">
            Last updated: {new Date(config.updated_at).toLocaleString()}
          </Typography>
        </Box>
      </Paper>
    </Container>
  );
}
