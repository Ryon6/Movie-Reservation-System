import { group } from 'k6';
import http from 'k6/http';
import { 
    BASE_URL, 
    headers, 
    getRandomUser, 
    login, 
    checkResponse,
    randomSleep 
} from './utils';

export const options = {
    scenarios: {
        real_user_simulation: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 100 },   // 2分钟内增加到100个用户
                { duration: '5m', target: 100 },   // 保持100个用户5分钟
                { duration: '2m', target: 200 },   // 2分钟内增加到200个用户
                { duration: '5m', target: 200 },   // 保持200个用户5分钟
                { duration: '2m', target: 0 },     // 2分钟内减少到0个用户
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<1000'],  // 95%的请求应该在1秒内完成
        errors: ['rate<0.1'],               // 错误率应该小于10%
    },
};

export default function () {
    group('Real User Simulation', function () {
        // 1. 用户登录
        const user = getRandomUser();
        type User = { username: string; password: string };
        const authHeaders = login((user as User).username, (user as User).password);
        if (!authHeaders) return;

        // 模拟用户思考时间
        randomSleep(2, 5);

        // 2. 浏览电影列表
        const movieListRes = http.get(`${BASE_URL}/api/v1/movies`, { headers: authHeaders });
        if (!checkResponse(movieListRes, 'Browse movies')) return;
        
        const movies = movieListRes.body ? JSON.parse(movieListRes.body as string).data : null;
        if (!movies || movies.length === 0) return;

        // 模拟用户思考时间
        randomSleep(3, 8);

        // 3. 随机选择一部电影，查看它的场次
        const randomMovie = movies[Math.floor(Math.random() * movies.length)];
        const showtimesRes = http.get(
            `${BASE_URL}/api/v1/showtimes?movieId=${randomMovie.id}`,
            { headers: authHeaders }
        );
        if (!checkResponse(showtimesRes, 'View showtimes')) return;

        const showtimes = showtimesRes.body ? JSON.parse(showtimesRes.body as string).data : null;
        if (!showtimes || showtimes.length === 0) return;

        // 模拟用户思考时间
        randomSleep(2, 6);

        // 4. 随机选择一个场次，查看座位图
        const randomShowtime = showtimes[Math.floor(Math.random() * showtimes.length)];
        const seatmapRes = http.get(
            `${BASE_URL}/api/v1/showtimes/${randomShowtime.id}/seatmap`,
            { headers: authHeaders }
        );
        if (!checkResponse(seatmapRes, 'View seat map')) return;

        const seatmap = seatmapRes.body ? JSON.parse(seatmapRes.body as string).data : null;
        if (!seatmap || !seatmap.seats || seatmap.seats.length === 0) return;

        // 模拟用户思考时间
        randomSleep(5, 10);

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
            { headers: authHeaders }
        );
        checkResponse(bookingRes, 'Book seat');
    });
} 