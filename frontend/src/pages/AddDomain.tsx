import React, { useState } from 'react';
import { Container, Typography, Box } from '@mui/material';
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
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        {t('addDomain')}
      </Typography>
      <Typography variant="body1" color="text.secondary" paragraph>
        {t('subtitle')}
      </Typography>

      <Box sx={{ mt: 3 }}>
        <SSLCheckForm onResult={handleResult} />
        <SSLResultDisplay result={result} />
      </Box>
    </Container>
  );
};

export default AddDomain;
