export interface UserCredential {
    username: string;
    password: string;
    email: string;
    role: string;
}

export const BASE_URL: string;
export const headers: Record<string, string>;
export const credentials: UserCredential[];

export function getRandomUser(): UserCredential;
export function login(username: string, password: string): Record<string, string> | null;
export function randomSleep(min?: number, max?: number): void;
export function checkResponse(res: any, checkName: string): boolean; 