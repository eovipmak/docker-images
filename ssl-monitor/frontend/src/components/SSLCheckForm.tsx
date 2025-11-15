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
import { addDomain, parseTarget } from '../services/api';
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
      const parsed = parseTarget(target);
      if (!parsed) {
        throw new Error('Invalid target');
      }

      const result = await addDomain(parsed.host, parsed.port);
      
      // Transform the response to match SSLCheckResponse format
      const sslCheckResponse: SSLCheckResponse = {
        status: result.status === 'success' ? 'success' : 'error',
        data: result.data?.check_result?.data,
      };
      
      onResult(sslCheckResponse);
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
    <Paper 
      elevation={0}
      sx={{ 
        p: 4,
        borderRadius: 3,
        background: 'white',
        border: '1px solid',
        borderColor: 'divider',
        mb: 3,
      }}
    >
      <Typography variant="h5" component="h2" gutterBottom fontWeight="bold">
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
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            }
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
          sx={{ 
            mt: 3,
            py: 1.5,
            borderRadius: 2,
            fontSize: '1rem',
            fontWeight: 600,
          }}
          aria-label={t('checkButton')}
        >
          {loading ? t('loading') : t('checkButton')}
        </Button>
      </Box>
    </Paper>
  );
};

export default SSLCheckForm;
