import { useContext } from 'react';
import { ThemeContext } from '../contexts/ThemeContext';
import lightLogo from '../assets/DcvixLogo.png';
import darkLogo from '../assets/DcvixLogoDarkTheme.png';

const Logo = ({ className = '', ...props }: { className?: string; style?: React.CSSProperties }) => {
    const { theme } = useContext(ThemeContext);

    return (
        <img
            src={theme === 'dark' ? darkLogo : lightLogo}
            alt="DCVIX Director"
            className={className}
            {...props}
        />
    );
};

export default Logo;
