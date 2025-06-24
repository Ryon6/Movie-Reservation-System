import { group } from 'k6';
import http from 'k6/http';
import { 
    BASE_URL, 
    getRandomUser, 
    login, 
    checkResponse 
} from './utils.ts';

// 测试场次ID（需要在运行前设置）
const TEST_SHOWTIME_ID = __ENV.SHOWTIME_ID || '101';

export const options = {
    scenarios: {
        rush_booking: {
            executor: 'ramping-arrival-rate',
            startRate: 0,
            timeUnit: '1s',
            preAllocatedVUs: 500,
            maxVUs: 1000,
            stages: [
                { duration: '30s', target: 0 },    // 等待所有VU就绪
                { duration: '1m', target: 500 },   // 1分钟内提升到500 RPS
                { duration: '3m', target: 500 },   // 维持500 RPS持续3分钟
                { duration: '30s', target: 0 },    // 30秒内降至0
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000'],  // 95%的请求应该在2秒内完成
        errors: ['rate<0.5'],               // 错误率可以较高，因为大量请求会因座位已被预订而失败
    },
};

// 初始化函数：获取座位图
export function setup() {
    const user = getRandomUser();
    type User = { username: string; password: string };
    const authHeaders = login((user as User).username, (user as User).password);
    if (!authHeaders) {
        console.error('Setup failed: Unable to login');
        return null;
    }

    const seatmapRes = http.get(
        `${BASE_URL}/api/v1/showtimes/${TEST_SHOWTIME_ID}/seatmap`,
        { headers: authHeaders }
    );

    if (seatmapRes.status !== 200) {
        console.error(`Setup failed: Unable to get seatmap - ${seatmapRes.status} ${seatmapRes.body}`);
        return null;
    }

    try {
        const seatmap = seatmapRes.body ? JSON.parse(seatmapRes.body as string).data : null;
        if (!seatmap || !seatmap.seats) {
            console.error('Setup failed: Invalid seatmap data');
            return null;
        }

        // 返回所有座位ID列表
        return {
            seatIds: seatmap.seats.map(seat => seat.id)
        };
    } catch (e) {
        console.error(`Setup failed: ${e.message}`);
        return null;
    }
}

export default function (data) {
    if (!data || !data.seatIds || data.seatIds.length === 0) {
        console.error('No seat data available');
        return;
    }

    group('Rush Booking Test', function () {
        // 1. 用户登录
        const user = getRandomUser();
        type User = { username: string; password: string };
        const authHeaders = login((user as User).username, (user as User).password);
        if (!authHeaders) return;

        // 2. 随机选择一个座位尝试预订
        const randomSeatId = data.seatIds[Math.floor(Math.random() * data.seatIds.length)];
        
        const bookingRes = http.post(
            `${BASE_URL}/api/v1/bookings`,
            JSON.stringify({
                showtimeId: TEST_SHOWTIME_ID,
                seatIds: [randomSeatId]
            }),
            { headers: authHeaders }
        );

        // 这里我们需要特别处理响应
        // 200: 预订成功
        // 409: 座位已被预订（期望的业务错误）
        // 401: 未授权（JWT问题）
        // 其他: 意外错误
        if (bookingRes.status !== 200 && bookingRes.status !== 409) {
            if (bookingRes.status === 401) {
                console.error('Booking failed: Unauthorized - JWT token might be invalid or expired');
            } else {
                console.error(`Booking failed with unexpected error: ${bookingRes.status} ${bookingRes.body}`);
            }
        }
    });
} 