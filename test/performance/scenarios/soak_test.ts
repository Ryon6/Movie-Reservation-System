import { group } from 'k6';
import http from 'k6/http';
import { 
    BASE_URL, 
    headers, 
    getRandomUser, 
    login, 
    checkResponse,
    randomSleep 
} from './utils.js';

export const options = {
    scenarios: {
        soak_test: {
            executor: 'constant-vus',
            vus: 100,              // 维持100个并发用户
            duration: '8h',        // 运行8小时
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<1000'],  // 95%的请求应该在1秒内完成
        errors: ['rate<0.05'],              // 错误率应该小于5%
        'http_req_duration{type:booking}': ['p(95)<2000'],  // 预订操作可以稍慢一些
    },
};

export default function () {
    group('Soak Test', function () {
        // 1. 用户登录
        const user = getRandomUser();
        type User = { username: string; password: string };
        const authHeaders = login((user as User).username, (user as User).password);
        if (!authHeaders) return;

        // 模拟用户思考时间（较长，因为是长期测试）
        randomSleep(5, 15);

        // 2. 浏览电影列表
        const movieListRes = http.get(`${BASE_URL}/api/v1/movies`, { 
            headers: authHeaders,
            tags: { type: 'browse' }
        });
        if (!checkResponse(movieListRes, 'Browse movies')) return;
        
        const movies = movieListRes.body ? JSON.parse(movieListRes.body as string).data : null;
        if (!movies || movies.length === 0) return;

        // 模拟用户思考时间
        randomSleep(5, 20);

        // 3. 随机选择一部电影，查看它的场次
        const randomMovie = movies[Math.floor(Math.random() * movies.length)];
        const showtimesRes = http.get(
            `${BASE_URL}/api/v1/showtimes?movieId=${randomMovie.id}`,
            { 
                headers: authHeaders,
                tags: { type: 'browse' }
            }
        );
        if (!checkResponse(showtimesRes, 'View showtimes')) return;

        const showtimes = showtimesRes.body ? JSON.parse(showtimesRes.body as string).data : null;
        if (!showtimes || showtimes.length === 0) return;

        // 模拟用户思考时间
        randomSleep(5, 15);

        // 4. 随机选择一个场次，查看座位图
        const randomShowtime = showtimes[Math.floor(Math.random() * showtimes.length)];
        const seatmapRes = http.get(
            `${BASE_URL}/api/v1/showtimes/${randomShowtime.id}/seatmap`,
            { 
                headers: authHeaders,
                tags: { type: 'browse' }
            }
        );
        if (!checkResponse(seatmapRes, 'View seat map')) return;

        const seatmap = seatmapRes.body ? JSON.parse(seatmapRes.body as string).data : null;
        if (!seatmap || !seatmap.seats || seatmap.seats.length === 0) return;

        // 模拟用户思考时间（较长，因为选座需要时间）
        randomSleep(10, 30);

        // 5. 随机选择一个可用座位进行预订
        const availableSeats = seatmap.seats.filter(seat => !seat.isBooked);
        if (availableSeats.length === 0) return;

        const randomSeat = availableSeats[Math.floor(Math.random() * availableSeats.length)];
        const bookingRes = http.post(
            `${BASE_URL}/api/v1/bookings`,
            JSON.stringify({
                showtimeId: randomShowtime.id,
                seatIds: [randomSeat.id]
            }),
            { 
                headers: authHeaders,
                tags: { type: 'booking' }
            }
        );
        checkResponse(bookingRes, 'Book seat');

        // 预订后较长休息，模拟真实用户行为
        randomSleep(30, 60);
    });
} 