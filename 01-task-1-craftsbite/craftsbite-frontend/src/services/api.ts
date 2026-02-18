import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';
import type { ApiErrorResponse } from '../types';
import { getToken, clearAuthData } from '../utils/storage';
import { API_BASE_URL } from '../utils/constants';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: 30000,
});

function buildApiError(error: AxiosError<ApiErrorResponse>): ApiErrorResponse {
    return {
        success: false,
        error: {
            code: error.response?.data?.error?.code || 'UNKNOWN_ERROR',
            message: error.response?.data?.error?.message || error.message || 'An unexpected error occurred',
            details: error.response?.data?.error?.details,
        },
    };
}

api.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = getToken();

        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
        }

        return config;
    },
    (error: AxiosError) => {
        return Promise.reject(error);
    }
);

api.interceptors.response.use(
    (response) => response,
    (error: AxiosError<ApiErrorResponse>) => {
        if (error.response?.status === 401) {
            clearAuthData();
            if (window.location.pathname !== '/login') {
                window.location.href = '/login';
            }
        }

        return Promise.reject(buildApiError(error));
    }
);

export default api;
export { buildApiError };