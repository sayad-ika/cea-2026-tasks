import React from 'react';

export interface FooterLink {
    label: string;
    href: string;
}

export interface FooterProps {
    brandName?: string;
    copyrightYear?: string;
    companyName?: string;
    links?: FooterLink[];
    className?: string;
}

const defaultLinks: FooterLink[] = [
    { label: 'Privacy', href: '#' },
    { label: 'Terms', href: '#' },
    { label: 'Support', href: '#' },
];

/**
 * Footer component for CraftsBite application
 * Displays brand info, copyright, and navigation links
 */
export const Footer: React.FC<FooterProps> = ({
    brandName = 'CraftsBite',
    copyrightYear = '2023',
    companyName = 'CraftsBite Inc.',
    links = defaultLinks,
    className = '',
}) => {
    return (
        <footer className={`w-full bg-background-light py-8 px-6 md:px-12 mt-auto border-t border-[#e6dccf]/40 ${className}`}>
            <div className="container mx-auto flex flex-col md:flex-row justify-between items-center gap-4 text-[#636E72]">
                {/* Brand Section */}
                <div className="flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary" style={{ fontSize: '1.25rem' }}>
                        restaurant
                    </span>
                    <span className="font-bold text-lg text-[#23170f]">{brandName}</span>
                    <span className="text-sm ml-2">
                        © {copyrightYear} {companyName}
                    </span>
                </div>

                {/* Links Section */}
                <nav>
                    <ul className="flex items-center gap-6 text-sm font-medium">
                        {links.map((link, index) => (
                            <li key={index}>
                                <a
                                    className="hover:text-primary transition-colors"
                                    href={link.href}
                                >
                                    {link.label}
                                </a>
                            </li>
                        ))}
                    </ul>
                </nav>
            </div>
        </footer>
    );
};
