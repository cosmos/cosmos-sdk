import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Brightness7Icon from '@mui/icons-material/Brightness7';
import Brightness4Icon from '@mui/icons-material/Brightness4';

interface Props {
    isMobile: boolean;
    darkMode: boolean;
    onChangeTheme: Function;
};

export default function App(props: Props) {
    const { darkMode, onChangeTheme } = props;
    return (
        <AppBar
            position="absolute"
            elevation={0}
            color="default"
            sx={{
                height: 55,
                backgroundColor: (theme) => darkMode ? theme.palette.grey[700] : theme.palette.grey[400],
            }}
        >
            <Toolbar
                variant="dense"
                disableGutters
                sx={{ px: 1, mt: 0.5 }}
            >
                <a href="/">
                    <Box
                        component="img"
                        src='/cosmos-wordmark.light.svg'
                        alt="Logo"
                        sx={{ height: 20, mt: 0.5, mr: 0.5 }}
                    />
                </a>
                <Typography variant="h6" color='lightgrey' sx={{ ml: 1 }}>
                    TM Load Tester
                </Typography>
                <Box sx={{ flexGrow: 1 }} />
                <IconButton
                    sx={{ ml: 1 }}
                    onClick={() => onChangeTheme()}
                    color="inherit"
                >
                    {
                        darkMode ?
                            <Brightness7Icon fontSize='small' /> :
                            <Brightness4Icon fontSize='small' />
                    }
                </IconButton>
            </Toolbar>
        </AppBar>
    );
}
