import React from 'react';
import { createTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

const DARK = 'dark';
const LIGHT = 'light';

export default function useColorMode() {
    const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)');
    const [mode, setMode] = React.useState(() => {
        const theme = localStorage.getItem('theme');
        if (theme === LIGHT || theme === DARK) return theme;
        return LIGHT;
    });

    const onChangeTheme = () => {
        setMode((m) => {
            const mode = m === LIGHT ? DARK : LIGHT;
            localStorage.setItem('theme', mode);
            return mode;
        });
    };

    const theme = React.useMemo(() => {
        const m = (prefersDarkMode || mode === DARK) ? DARK : LIGHT;
        return createTheme({
            palette: {
                mode: m,
                contrastThreshold: 3,
                primary: {
                    main: '#4d73e1',
                },
            },
        });
    }, [prefersDarkMode, mode]);

    return { theme, onChangeTheme };
}
