import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export const Register: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [nameError, setNameError] = useState("");
  const [emailError, setEmailError] = useState("");
  const [passwordError, setPasswordError] = useState("");

  // Redirect if already authenticated
  React.useEffect(() => {
    if (isAuthenticated) {
      navigate("/home");
    }
  }, [isAuthenticated, navigate]);

  // Name validation
  const validateName = (name: string): boolean => {
    if (!name.trim()) {
      setNameError("Full name is required");
      return false;
    }
    if (name.trim().length < 2) {
      setNameError("Name must be at least 2 characters");
      return false;
    }
    setNameError("");
    return true;
  };

  // Email validation
  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!email) {
      setEmailError("Email is required");
      return false;
    }
    if (!emailRegex.test(email)) {
      setEmailError("Please enter a valid email address");
      return false;
    }
    setEmailError("");
    return true;
  };

  // Password validation
  const validatePassword = (password: string): boolean => {
    if (!password) {
      setPasswordError("Password is required");
      return false;
    }
    if (password.length < 8) {
      setPasswordError("Password must be at least 8 characters");
      return false;
    }
    setPasswordError("");
    return true;
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    // Validate fields
    const isNameValid = validateName(name);
    const isEmailValid = validateEmail(email);
    const isPasswordValid = validatePassword(password);

    if (!isNameValid || !isEmailValid || !isPasswordValid) {
      return;
    }

    setIsLoading(true);

    try {
      // Import authService dynamically to avoid circular dependency
      const { register } = await import("../services/authService");
      await register(name, email, "employee", password);

      // Show success message and redirect to login
      navigate("/login");
    } catch (err: any) {
      setError(err?.error?.message || "Registration failed. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="font-display bg-[#FFF5E6] text-[#4a4a4a] min-h-screen flex flex-col overflow-hidden relative">
      {/* Floating Food Icons */}
      <div className="absolute top-10 left-10 md:top-20 md:left-32 text-primary/20 floating-element z-0">
        <span className="material-symbols-outlined text-8xl md:text-9xl">
          lunch_dining
        </span>
      </div>
      <div className="absolute bottom-20 right-10 md:bottom-32 md:right-32 text-primary/20 floating-element-delay z-0">
        <span className="material-symbols-outlined text-8xl md:text-9xl">
          bakery_dining
        </span>
      </div>
      <div className="absolute top-1/4 right-20 text-orange-200/30 floating-element z-0 hidden md:block">
        <span className="material-symbols-outlined text-7xl">icecream</span>
      </div>
      <div className="absolute bottom-1/4 left-20 text-orange-200/30 floating-element-delay z-0 hidden md:block">
        <span className="material-symbols-outlined text-7xl">local_pizza</span>
      </div>

      {/* Main Content */}
      <main className="flex-grow flex items-center justify-center relative z-20 px-4 py-16">
        <div className="w-full max-w-lg bg-[#FFFDF5] rounded-[40px] shadow-clay p-8 md:p-12 border border-white/60 relative overflow-hidden">
          {/* Gradient Blurs */}
          <div className="absolute -top-20 -right-20 w-64 h-64 bg-orange-100 rounded-full blur-3xl opacity-50 pointer-events-none"></div>
          <div className="absolute -bottom-20 -left-20 w-64 h-64 bg-yellow-100 rounded-full blur-3xl opacity-50 pointer-events-none"></div>

          <div className="relative z-10">
            {/* Title */}
            <div className="text-center mb-10">
              <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white flex items-center justify-center shadow-clay-button transform -rotate-6">
                <span className="material-symbols-outlined text-3xl">
                  restaurant
                </span>
              </div>
              <h2 className="text-3xl font-black text-[#23170f] tracking-tight mb-2">
                Join CraftsBite
              </h2>
              <p className="text-[#8c705f] font-medium">
                Delicious meals managed effortlessly.
              </p>
            </div>

            {/* Registration Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Global Error Message */}
              {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-2xl text-sm">
                  {error}
                </div>
              )}

              {/* Full Name Field */}
              <div className="space-y-2">
                <label
                  className="block text-sm font-bold text-[#23170f] ml-1"
                  htmlFor="fullname"
                >
                  Full Name
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-[#8c705f]/70 group-focus-within:text-[#fa8c47] transition-colors">
                      person
                    </span>
                  </div>
                  <input
                    className="w-full pl-12 pr-4 py-3.5 bg-[#FFFDF5] border border-[#fa8c47]/20 rounded-2xl shadow-clay-inset focus:ring-2 focus:ring-[#fa8c47]/50 focus:border-[#fa8c47] text-[#4a4a4a] placeholder-[#8c705f]/40 outline-none transition-all duration-200"
                    id="fullname"
                    type="text"
                    placeholder="e.g. Alex Johnson"
                    value={name}
                    onChange={(e) => {
                      setName(e.target.value);
                      if (nameError) validateName(e.target.value);
                    }}
                    onBlur={() => validateName(name)}
                  />
                </div>
                {nameError && (
                  <p className="text-red-600 text-sm ml-1">{nameError}</p>
                )}
              </div>

              {/* Email Field */}
              <div className="space-y-2">
                <label
                  className="block text-sm font-bold text-[#23170f] ml-1"
                  htmlFor="email"
                >
                  Work Email
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-[#8c705f]/70 group-focus-within:text-[#fa8c47] transition-colors">
                      mail
                    </span>
                  </div>
                  <input
                    className="w-full pl-12 pr-4 py-3.5 bg-[#FFFDF5] border border-[#fa8c47]/20 rounded-2xl shadow-clay-inset focus:ring-2 focus:ring-[#fa8c47]/50 focus:border-[#fa8c47] text-[#4a4a4a] placeholder-[#8c705f]/40 outline-none transition-all duration-200"
                    id="email"
                    type="email"
                    placeholder="alex@company.com"
                    value={email}
                    onChange={(e) => {
                      setEmail(e.target.value);
                      if (emailError) validateEmail(e.target.value);
                    }}
                    onBlur={() => validateEmail(email)}
                  />
                </div>
                {emailError && (
                  <p className="text-red-600 text-sm ml-1">{emailError}</p>
                )}
              </div>

              {/* Role Field */}
              {/* <div className="space-y-2">
                <label
                  className="block text-sm font-bold text-[#23170f] ml-1"
                  htmlFor="role"
                >
                  Role
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-[#8c705f]/70 group-focus-within:text-[#fa8c47] transition-colors">
                      business_center
                    </span>
                  </div>
                  <select
                    className="w-full pl-12 pr-10 py-3.5 bg-[#FFFDF5] border border-[#fa8c47]/20 rounded-2xl shadow-clay-inset focus:ring-2 focus:ring-[#fa8c47]/50 focus:border-[#fa8c47] text-[#4a4a4a] outline-none transition-all duration-200 appearance-none cursor-pointer"
                    id="role"
                    value={role}
                    onChange={(e) => setRole(e.target.value as UserRole)}
                  >
                    <option value="employee">Employee</option>
                    <option value="team_lead">Team Lead</option>
                    <option value="admin">Admin</option>
                    <option value="logistics">Logistics</option>
                  </select>
                  <div className="absolute inset-y-0 right-0 pr-4 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-[#8c705f]/50">
                      expand_more
                    </span>
                  </div>
                </div>
              </div> */}

              {/* Password Field */}
              <div className="space-y-2">
                <label
                  className="block text-sm font-bold text-[#23170f] ml-1"
                  htmlFor="password"
                >
                  Create Password
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-[#8c705f]/70 group-focus-within:text-[#fa8c47] transition-colors">
                      lock
                    </span>
                  </div>
                  <input
                    className="w-full pl-12 pr-4 py-3.5 bg-[#FFFDF5] border border-[#fa8c47]/20 rounded-2xl shadow-clay-inset focus:ring-2 focus:ring-[#fa8c47]/50 focus:border-[#fa8c47] text-[#4a4a4a] placeholder-[#8c705f]/40 outline-none transition-all duration-200"
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
                {passwordError && (
                  <p className="text-red-600 text-sm ml-1">{passwordError}</p>
                )}
                <p className="text-xs text-[#8c705f] ml-1 mt-1">
                  Must be at least 8 characters
                </p>
              </div>

              {/* Submit Button */}
              <button
                className="w-full py-4 mt-4 rounded-2xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold text-lg shadow-[8px_8px_16px_#e6dccf,-8px_-8px_16px_#ffffff] hover:scale-[1.02] hover:shadow-[10px_10px_20px_#e6dccf,-10px_-10px_20px_#ffffff] active:scale-[0.98] active:shadow-[inset_4px_4px_8px_#c26629,inset_-4px_-4px_8px_#ffb275] transition-all duration-200 flex items-center justify-center gap-2 group disabled:opacity-50 disabled:cursor-not-allowed"
                type="submit"
                disabled={isLoading}
              >
                {isLoading ? (
                  <>
                    <span className="animate-spin material-symbols-outlined">
                      progress_activity
                    </span>
                    <span>Creating Account...</span>
                  </>
                ) : (
                  <>
                    <span>Create Account</span>
                    <span className="material-symbols-outlined group-hover:translate-x-1 transition-transform">
                      arrow_forward
                    </span>
                  </>
                )}
              </button>
            </form>

            {/* Back to Login Link */}
            <div className="mt-8 pt-6 border-t border-[#fa8c47]/10 text-center">
              <p className="text-[#8c705f] text-sm mb-4">
                Already have an account?
              </p>
              <Link
                className="inline-flex items-center gap-2 text-[#fa8c47] font-bold hover:text-[#e57a36] transition-colors group"
                to="/login"
              >
                <span className="material-symbols-outlined text-lg group-hover:-translate-x-1 transition-transform">
                  arrow_back
                </span>
                Back to Login
              </Link>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="w-full py-4 text-center text-xs text-[#8c705f]/60 relative z-10">
        © 2023 CraftsBite. All rights reserved.
      </footer>

      {/* Floating Animation Styles */}
      <style>{`
                @keyframes float {
                    0% { transform: translateY(0px) rotate(0deg); }
                    50% { transform: translateY(-20px) rotate(5deg); }
                    100% { transform: translateY(0px) rotate(0deg); }
                }
                .floating-element {
                    animation: float 6s ease-in-out infinite;
                }
                .floating-element-delay {
                    animation: float 7s ease-in-out infinite reverse;
                }
            `}</style>
    </div>
  );
};
