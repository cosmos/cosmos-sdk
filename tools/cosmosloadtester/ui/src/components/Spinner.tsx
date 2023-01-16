import { FC } from 'react';
import CircularProgress from '@mui/material/CircularProgress';

interface Props {
  loading: boolean;
};

export const Spinner: FC<Props> = ({ loading }) => {
    const styles: {[key: string]: number|string} = {
        position: 'absolute',
        top: '0%',
        left: '0%',
        right: '0%',
        bottom: '0%',
        visibility: 'hidden',
        backgroundColor: '#ffffff00',
        transition: 'background-color 200ms ease-in-out',
        zIndex: -1,
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center'
    };

    if (loading) {
        styles['visibility'] = 'visible';
        styles['backgroundColor'] = '#ffffffd9';
        styles['zIndex'] = 999;
        styles['opacity'] = 0.5;
    }
    
    return (
        <div style={styles}>
            <CircularProgress size={100} color='info' />
        </div>
    );
};
