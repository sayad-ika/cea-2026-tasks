// API Response Types

export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    message?: string;
}

export interface ApiError {
    code: string;
    message: string;
    details?: Array<{
        field: string;
        message: string;
    }>;
}

export interface ApiErrorResponse {
    success: false;
    error: ApiError;
}

export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    limit: number;
    totalPages: number;
}

export interface UserListParams {
    role?: string;
    active?: boolean;
    page?: number;
    limit?: number;
}

export interface CreateUserRequest {
    email: string;
    name: string;
    password: string;
    role: string;
}

export interface UpdateUserRequest {
    email?: string;
    name?: string;
    role?: string;
    active?: boolean;
}
