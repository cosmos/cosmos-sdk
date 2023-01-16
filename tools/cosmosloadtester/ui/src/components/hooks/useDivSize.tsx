import { useState, useEffect, RefObject } from 'react';

export const useDivSize = (divRef: RefObject<HTMLDivElement>) => {
    const [width, setWidth] = useState(0);
    const [height, setHeight] = useState(0);

    useEffect(() => {
        try {
            if (divRef.current != null) {
                let screenRatio = 4 / 9;
                /**
                  less than 320px: Very small devices
                  320px — 480px: Mobile devices
                  481px — 768px: iPads, Tablets
                  769px — 1024px: Small screens, laptops
                  1025px — 1200px: Desktops, large screens
                  1201px and more —  Extra large screens, TV
                */
                const w = divRef.current.clientWidth;
                if (w < 320) screenRatio = 6 / 9;
                if (w >= 320 && w <= 480) screenRatio = 4 / 6;
                if (w >= 481 && w <= 768) screenRatio = 4 / 9;
                if (w >= 769 && w <= 1024) screenRatio = 3 / 8;
                if (w >= 1025 && w <= 1200) screenRatio = 3 / 9;
                if (w >= 1201) screenRatio = 2 / 8;
                const h = screenRatio * w;
                setWidth(w);
                setHeight(h);
            }
        } catch (e) { }
    }, [divRef]);

    useEffect(() => {
        const handleResize = () => {
            if (divRef.current == null) return;
            const w = divRef.current.clientWidth;
            const h = divRef.current.clientHeight;
            // do nothing if the width does not change
            if (w === width) return;
            setWidth(w);
            setHeight(h);
        };
        window.addEventListener('resize', handleResize);
        return () => {
            window.removeEventListener('resize', handleResize);
        };
    });
    return { width, height };
}
