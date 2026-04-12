import { useState, ReactNode } from 'react';

interface TooltipProps {
    text: string;
    children: ReactNode;
    position?: 'top' | 'right' | 'bottom' | 'left';
}

export default function Tooltip({ text, children, position = 'right' }: TooltipProps) {
    const [show, setShow] = useState(false);

    const positionClasses: Record<string, string> = {
        top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
        right: 'left-full top-1/2 -translate-y-1/2 ml-2',
        bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
        left: 'right-full top-1/2 -translate-y-1/2 mr-2',
    };

    return (
        <span
            className="relative inline-flex"
            onMouseEnter={() => setShow(true)}
            onMouseLeave={() => setShow(false)}
        >
            {children}
            {show && (
                <span
                    className={`absolute z-50 px-3 py-2 text-xs text-white bg-gray-800 rounded-lg shadow-lg whitespace-normal max-w-xs ${positionClasses[position]}`}
                >
                    {text}
                </span>
            )}
        </span>
    );
}
