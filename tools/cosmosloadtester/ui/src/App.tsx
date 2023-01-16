import React from 'react';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Container from '@mui/material/Container';
import useMediaQuery from '@mui/material/useMediaQuery';

import useColorMode from './components/hooks/useColorMode';
import Footer from './components/foooter/Footer';
import AppToBar from './components/navigation/AppTopBar';

import LoadTester from './components/load-tester/LoadTester';

export default function App() {
  const { theme, onChangeTheme } = useColorMode();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const darkMode = theme.palette.mode === 'dark';

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppToBar isMobile={isMobile} darkMode={darkMode} onChangeTheme={onChangeTheme} />
      <Box
        component="main"
        sx={{
          bgcolor: ({ palette: p }) => p.mode === 'dark' ? p.grey[800] : p.grey[300],
          height: '100vh',
          width: '100%',
          overflow: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <Toolbar /> 
        <Container maxWidth="lg" sx={{ flexGrow: 1 }}>
          <LoadTester />
        </Container>
        <Footer />
      </Box>
    </ThemeProvider>
  );
}
