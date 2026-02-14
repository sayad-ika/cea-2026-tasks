// Axios API Instance Configuration

import axios, { AxiosError } from 'axios';
import type { ApiErrorResponse } from '../types';
import { clearAuthData } from '../utils/storage';
import { API_BASE_URL } from '../utils/constants';

// Create axios instance
const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: 30000, // 30 seconds
    withCredentials: true, // Enable cookies
});

// Request interceptor - attach JWT token
// api.interceptors.request.use(
//     (config: InternalAxiosRequestConfig) => {
//         const token = getToken();

//         if (token && config.headers) {
//             config.headers.Authorization = `Bearer ${token}`;
//         }

//         return config;
//     },
//     (error: AxiosError) => {
//         return Promise.reject(error);
//     }
// );

// Response interceptor - handle errors
api.interceptors.response.use(
    (response) => {
        return response;
    },
    (error: AxiosError<ApiErrorResponse>) => {
        // Handle 401 Unauthorized - clear auth data and redirect to login
        if (error.response?.status === 401) {
            clearAuthData();

            // Only redirect if not already on login page
            if (window.location.pathname !== '/login') {
                window.location.href = '/login';
            }
        }

        // Format error response
        const errorResponse: ApiErrorResponse = {
            success: false,
            error: {
                code: error.response?.data?.error?.code || 'UNKNOWN_ERROR',
                message: error.response?.data?.error?.message || error.message || 'An unexpected error occurred',
                details: error.response?.data?.error?.details,
            },
        };

        return Promise.reject(errorResponse);
    }
);

export default api;
