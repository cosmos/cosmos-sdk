import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import Box from '@mui/material/Box';

const Brand = () => (
    <Box sx={{
        display: 'flex',
        alignItems: 'center',
        marginTop: 'auto',
    }}
    >
        <Link
            href="https://energy.orijtech.com"
            target="_blank"
            sx={{ textDecoration: 'none', color: 'gray' }}
        >
            <Box component="img" src="/orijtech-logo+wordmark.svg" sx={{ height: 40 }} />
        </Link>
    </Box>
);

export default function Footer() {
    return (
        <Grid container spacing={2} sx={{ px: 6, py: 4, mt: 'auto' }}>
            <Grid item xs={6}>
                <Brand />
            </Grid>
        </Grid>
    );
};
