import React, { useState } from 'react';
import {
  Box,
  TextField,
  Button,
  CircularProgress,
  Typography,
  Paper,
} from '@mui/material';
import { Search as SearchIcon } from '@mui/icons-material';
import { useLanguage } from '../hooks/useLanguage';
import { checkSSL } from '../services/api';
import type { SSLCheckResponse } from '../types';

interface SSLCheckFormProps {
  onResult: (result: SSLCheckResponse) => void;
}

const SSLCheckForm: React.FC<SSLCheckFormProps> = ({ onResult }) => {
  const { t } = useLanguage();
  const [target, setTarget] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!target.trim()) {
      setError(t('provideDomain'));
      return;
    }

    setLoading(true);

    try {
      const result = await checkSSL(target);
      onResult(result);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t('checkFailed');
      setError(errorMessage);
      onResult({
        status: 'error',
        error: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper elevation={2} sx={{ p: 3 }}>
      <Typography variant="h5" component="h2" gutterBottom>
        {t('formTitle')}
      </Typography>
      <Box component="form" onSubmit={handleSubmit} noValidate>
        <TextField
          fullWidth
          label={t('targetLabel')}
          placeholder={t('targetPlaceholder')}
          value={target}
          onChange={(e) => setTarget(e.target.value)}
          error={!!error}
          helperText={error || t('helpText')}
          disabled={loading}
          margin="normal"
          inputProps={{
            'aria-label': t('targetLabel'),
            'aria-required': 'true',
          }}
        />
        <Button
          type="submit"
          variant="contained"
          color="primary"
          size="large"
          startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <SearchIcon />}
          disabled={loading}
          fullWidth
          sx={{ mt: 2 }}
          aria-label={t('checkButton')}
        >
          {loading ? t('loading') : t('checkButton')}
        </Button>
      </Box>
    </Paper>
  );
};

export default SSLCheckForm;
