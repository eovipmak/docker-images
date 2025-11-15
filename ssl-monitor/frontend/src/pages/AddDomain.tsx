import React, { useState } from 'react';
import { Container, Typography, Box, Fade } from '@mui/material';
import { AddCircle as AddCircleIcon } from '@mui/icons-material';
import { useLanguage } from '../hooks/useLanguage';
import SSLCheckForm from '../components/SSLCheckForm';
import SSLResultDisplay from '../components/SSLResultDisplay';
import type { SSLCheckResponse } from '../types';

const AddDomain: React.FC = () => {
  const { t } = useLanguage();
  const [result, setResult] = useState<SSLCheckResponse | null>(null);

  const handleResult = (res: SSLCheckResponse) => {
    setResult(res);
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%)',
        pb: 4,
      }}
    >
      <Container maxWidth="lg" sx={{ pt: 4 }}>
        <Fade in={true} timeout={500}>
          <Box mb={4}>
            <Box display="flex" alignItems="center" gap={2} mb={2}>
              <Box
                sx={{
                  p: 1.5,
                  borderRadius: 2,
                  background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                  color: 'white',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <AddCircleIcon sx={{ fontSize: 28 }} />
              </Box>
              <Box>
                <Typography 
                  variant="h3" 
                  component="h1" 
                  fontWeight="bold"
                  sx={{
                    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                    backgroundClip: 'text',
                  }}
                >
                  {t('addDomain')}
                </Typography>
                <Typography variant="body1" color="text.secondary">
                  {t('subtitle')}
                </Typography>
              </Box>
            </Box>
          </Box>
        </Fade>

        <Box sx={{ mt: 3 }}>
          <SSLCheckForm onResult={handleResult} />
          <SSLResultDisplay result={result} />
        </Box>
      </Container>
    </Box>
  );
};

export default AddDomain;
