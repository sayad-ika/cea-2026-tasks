import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export const Login: React.FC = () => {
    const navigate = useNavigate();
    const { login, isLoading, isAuthenticated } = useAuth();

    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [emailError, setEmailError] = useState('');
    const [passwordError, setPasswordError] = useState('');

    // Redirect if already authenticated
    React.useEffect(() => {
        if (isAuthenticated) {
            navigate('/home');
        }
    }, [isAuthenticated, navigate]);

    // Email validation
    const validateEmail = (email: string): boolean => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!email) {
            setEmailError('Email is required');
            return false;
        }
        if (!emailRegex.test(email)) {
            setEmailError('Please enter a valid email address');
            return false;
        }
        setEmailError('');
        return true;
    };

    // Password validation
    const validatePassword = (password: string): boolean => {
        if (!password) {
            setPasswordError('Password is required');
            return false;
        }
        if (password.length < 6) {
            setPasswordError('Password must be at least 6 characters');
            return false;
        }
        setPasswordError('');
        return true;
    };

    // Handle form submission
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        // Validate fields
        const isEmailValid = validateEmail(email);
        const isPasswordValid = validatePassword(password);

        if (!isEmailValid || !isPasswordValid) {
            return;
        }

        try {
            await login(email, password);
            // Redirect to dashboard on successful login
            navigate('/home');
        } catch (err: any) {
            setError(err?.error?.message || 'Invalid email or password. Please try again.');
        }
    };

    return (
        <div className="font-display bg-[#FFF5E6] text-[#4a4a4a] min-h-screen flex flex-col overflow-hidden relative">
            {/* Floating Food Emojis */}
            <div className="absolute top-20 left-20 text-6xl opacity-80 animate-float blur-[1px]">🍕</div>
            <div className="absolute bottom-32 left-32 text-5xl opacity-60 animate-float-delayed blur-[1px]">🥗</div>
            <div className="absolute top-32 right-32 text-6xl opacity-80 animate-float-delayed blur-[1px]">🍔</div>
            <div className="absolute bottom-20 right-20 text-5xl opacity-60 animate-float blur-[1px]">🍩</div>
            <div className="absolute top-1/2 left-10 text-4xl opacity-40 animate-float blur-[2px]">🥑</div>
            <div className="absolute top-1/3 right-10 text-4xl opacity-40 animate-float-delayed blur-[2px]">🌮</div>

            {/* Header */}
            <header className="w-full px-6 py-4 md:px-12 flex justify-between items-center z-10 absolute top-0">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-[#fa8c47] text-white flex items-center justify-center shadow-clay-button">
                        <span className="material-symbols-outlined">restaurant_menu</span>
                    </div>
                    <h1 className="text-2xl font-black tracking-tight text-[#23170f]">CraftsBite</h1>
                </div>
            </header>

            {/* Main Content */}
            <main className="flex-grow flex items-center justify-center relative z-20 px-4">
                <div className="w-full max-w-md bg-[#FFFDF5] rounded-[40px] shadow-clay-card p-8 md:p-12 border border-white/60 relative overflow-hidden">
                    {/* Gradient Blurs */}
                    <div className="absolute -top-20 -right-20 w-64 h-64 bg-orange-100 rounded-full blur-3xl opacity-50 pointer-events-none"></div>
                    <div className="absolute -bottom-20 -left-20 w-64 h-64 bg-yellow-100 rounded-full blur-3xl opacity-50 pointer-events-none"></div>

                    <div className="relative z-10">
                        {/* Title */}
                        <div className="text-center mb-10">
                            <h2 className="text-3xl font-black text-[#23170f] mb-2 tracking-tight">Welcome Back!</h2>
                            <p className="text-[#8c705f] font-medium">Please enter your details.</p>
                        </div>

                        {/* Login Form */}
                        <form onSubmit={handleSubmit} className="space-y-6">
                            {/* Global Error Message */}
                            {error && (
                                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                                    {error}
                                </div>
                            )}

                            {/* Email Field */}
                            <div className="space-y-2">
                                <label className="block text-sm font-bold text-[#4a4a4a] ml-1" htmlFor="email">
                                    Email
                                </label>
                                <div className="relative group">
                                    <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                                        <span className="material-symbols-outlined text-[#8c705f]/70 group-focus-within:text-[#fa8c47] transition-colors">
                                            mail
                                        </span>
                                    </div>
                                    <input
                                        className="w-full pl-12 pr-4 py-4 rounded-2xl bg-[#FFFDF5] border-none shadow-clay-inset focus:ring-2 focus:ring-[#fa8c47]/50 focus:shadow-inner text-[#4a4a4a] placeholder-[#8c705f]/50 outline-none transition-all duration-200"
                                        id="email"
                                        type="email"
                                        placeholder="name@company.com"
                                        value={email}
                                        onChange={(e) => {
                                            setEmail(e.target.value);
                                            if (emailError) validateEmail(e.target.value);
                                        }}
                                        onBlur={() => validateEmail(email)}
                                    />
                                </div>
                                {emailError && <p className="text-red-600 text-sm ml-1">{emailError}</p>}
                            </div>

                            {/* Password Field */}
                            <div className="space-y-2">
                                <label className="block text-sm font-bold text-[#4a4a4a] ml-1" htmlFor="password">
                                    Password
                                </label>
                                <div className="relative group">
                                    <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                                        <span className="material-symbols-outlined text-[#8c705f]/70 group-focus-within:text-[#fa8c47] transition-colors">
                                            lock
                                        </span>
                                    </div>
                                    <input
                                        className="w-full pl-12 pr-4 py-4 rounded-2xl bg-[#FFFDF5] border-none shadow-clay-inset focus:ring-2 focus:ring-[#fa8c47]/50 focus:shadow-inner text-[#4a4a4a] placeholder-[#8c705f]/50 outline-none transition-all duration-200"
                                        id="password"
                                        type="password"
                                        placeholder="••••••••"
                                        value={password}
                                        onChange={(e) => {
                                            setPassword(e.target.value);
                                            if (passwordError) validatePassword(e.target.value);
                                        }}
                                        onBlur={() => validatePassword(password)}
                                    />
                                </div>
                                {passwordError && <p className="text-red-600 text-sm ml-1">{passwordError}</p>}

                                <div className="flex justify-end pt-1">
                                    <a className="text-sm font-bold text-[#fa8c47] hover:text-[#e57a36] transition-colors" href="#">
                                        Forgot Password?
                                    </a>
                                </div>
                            </div>

                            {/* Submit Button */}
                            <button
                                className="w-full py-4 mt-4 rounded-2xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold text-lg shadow-[8px_8px_16px_#e6dccf,-8px_-8px_16px_#ffffff] hover:scale-[1.02] hover:shadow-[10px_10px_20px_#e6dccf,-10px_-10px_20px_#ffffff] active:scale-[0.98] active:shadow-[inset_4px_4px_8px_#c26629,inset_-4px_-4px_8px_#ffb275] transition-all duration-200 flex items-center justify-center gap-2 group disabled:opacity-50 disabled:cursor-not-allowed"
                                type="submit"
                                disabled={isLoading}
                            >
                                {isLoading ? (
                                    <>
                                        <span className="animate-spin material-symbols-outlined">progress_activity</span>
                                        <span>Signing In...</span>
                                    </>
                                ) : (
                                    <>
                                        <span>Sign In</span>
                                        <span className="material-symbols-outlined group-hover:translate-x-1 transition-transform">
                                            arrow_forward
                                        </span>
                                    </>
                                )}
                            </button>
                        </form>

                    </div>
                </div>
            </main>

            {/* Footer */}
            <footer className="w-full py-6 text-center relative z-10 text-[#8c705f]/60 text-sm">
                <p>© 2026 CraftsBite. All rights reserved.</p>
            </footer>

            {/* CSS Animations */}
            <style>{`
        @keyframes float {
          0% { transform: translateY(0px); }
          50% { transform: translateY(-20px); }
          100% { transform: translateY(0px); }
        }
        .animate-float {
          animation: float 6s ease-in-out infinite;
        }
        .animate-float-delayed {
          animation: float 6s ease-in-out infinite;
          animation-delay: 3s;
        }
        .shadow-clay-card {
          box-shadow: 30px 30px 60px #e6dccf, -30px -30px 60px #ffffff;
        }
        .shadow-clay-inset {
          box-shadow: inset 6px 6px 12px #e6dccf, inset -6px -6px 12px #ffffff;
        }
        .shadow-clay-button {
          box-shadow: 8px 8px 16px #e6dccf, -8px -8px 16px #ffffff;
        }
      `}</style>
        </div>
    );
};
