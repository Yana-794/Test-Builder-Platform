/**
 * Нормализованная ошибка обращения к апи-шлюзу
 */

export class ApiError extends Error {
    readonly status:  number;
    readonly code?: string;
    readonly details?: unknown;
    constructor(message: string, status: number, options?: {code?: string; details?: unknown }) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.code = options?.code;
        this.details = options?.details;
    }
    get isNetworkError(): boolean{
        return this.status === 0
    }
}

