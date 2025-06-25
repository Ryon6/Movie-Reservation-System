import { SharedArray } from 'k6/data';
import { check, sleep } from 'k6';
import http from 'k6/http';
import { Rate } from 'k6/metrics';

// 自定义指标
export const errorRate = new Rate('errors');

// API 基础URL
export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// 加载用户凭证
export const credentials = new SharedArray('users', function () {
    return JSON.parse(open('../data/user_credentials.json'));
});

// 通用请求头
export const headers = {
    'Content-Type': 'application/json',
};

// 获取随机用户凭证
export function getRandomUser() {
    return credentials[Math.floor(Math.random() * credentials.length)];
}

// 用户登录并获取token
export function login(username, password) {
    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
        username: username,
        password: password
    }), { headers });

    const success = check(loginRes, {
        'login successful': (r) => r.status === 200,
        'has token': (r) => {
            try {
                const body = r.body ? JSON.parse(r.body.toString()) : null;
                return body?.token;
            } catch (e) {
                return false;
            }
        },
    });

    if (!success) {
        console.error(`Login failed: ${loginRes.body}`);
        return null;
    }

    try {
        const body = loginRes.body ? JSON.parse(loginRes.body.toString()) : null;
        const token = body?.token;
        if (!token) return null;

        return {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        };
    } catch (e) {
        console.error(`Failed to parse login response: ${e.message}`);
        return null;
    }
}

// 随机延迟函数（模拟用户思考时间）
export function randomSleep(min = 1, max = 5) {
    sleep(Math.random() * (max - min) + min);
}

// 检查响应状态
export function checkResponse(res, checkName, status = 200) {
    const success = check(res, {
        [`${checkName} status is ${status}`]: (r) => r.status === status,
        [`${checkName} response is valid`]: (r) => {
            if (!r.body) return false;
            if (typeof r.body === 'string') return r.body.length > 0;
            return r.body.byteLength > 0;
        },
    });

    if (!success) {
        errorRate.add(1);
        if (res.status === 401) {
            console.error(`${checkName} failed: Unauthorized - JWT token might be invalid or expired`);
        } else {
            console.error(`${checkName} failed: ${res.status} ${res.body}`);
        }
    }
    return success;
} 