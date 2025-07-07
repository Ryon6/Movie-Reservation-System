import { check, group } from 'k6';
import http from 'k6/http';
// @ts-ignore
import { URLSearchParams } from 'https://jslib.k6.io/url/1.0.0/index.js';

import {
    BASE_URL,
    getRandomUser,
    login,
    checkResponse,
    randomSleep
} from './utils.js';

export const options = {
    scenarios: {
        real_user_simulation: {
            executor: 'constant-vus',
            vus: 1000,
            duration: '1m',
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<1000'],
        errors: ['rate<0.1'],
    },
};

const genres = ["动作", "冒险", "喜剧", "剧情", "恐怖", "科幻", "动画", "纪录片"];

// 生成电影查询参数
function generateMovieQueryParams() {
    const params = new URLSearchParams();

    // 随机决定使用哪些参数
    const useGenre = Math.random() < 0.01;      // 1%概率使用类型搜索
    const useYear = Math.random() < 0.01;       // 1%概率使用年份搜索

    if (useGenre) {
        const randomGenre = genres[Math.floor(Math.random() * genres.length)];
        params.set('genre_name', encodeURIComponent(randomGenre));
    }

    if (useYear) {
        // 生成2000-2024之间的随机年份
        const randomYear = Math.floor(Math.random() * 25) + 2000;
        params.set('release_year', randomYear.toString());
    }

    // 添加分页参数
    params.set('page', '1');
    params.set('page_size', '20');

    return params.toString();
}

export default function () {
    group('Real User Simulation', function () {
        // 获取用户
        const user = getRandomUser();
        randomSleep(0, 1);
        const authHeaders = login(user['username'], user['password']);
        if (!authHeaders) return;

        // 模拟用户思考时间
        randomSleep(1, 3);

        // 1. 用户查询电影列表
        const movieListRes = http.get(
            `${BASE_URL}/api/v1/movies?${generateMovieQueryParams()}`,
            { headers: authHeaders }
        );
        checkResponse(movieListRes, 'Get movies list');

        if (movieListRes.status !== 200) return;
        const movies = movieListRes?.json()?.['movies'] ?? [];
        if (movies.length === 0) return;

        // 2. 用户挑选一部喜爱的电影 (随机选择一部)
        const favoriteMovie = movies[Math.floor(Math.random() * movies.length)];
        const movieId = favoriteMovie['id'];
        randomSleep(2, 5); // 模拟用户查看和选择电影的思考时间

        // 3. 用户通过电影ID查询该电影的放映计划
        const showtimeListRes = http.get(
            `${BASE_URL}/api/v1/showtimes?movie_id=${movieId}&page=1&page_size=20`,
            { headers: authHeaders }
        );
        checkResponse(showtimeListRes, `Get showtimes for movie`);

        if (showtimeListRes.status !== 200) return;

        const showtimes = showtimeListRes?.json()?.['showtimes'] ?? [];
        if (showtimes.length === 0) return;

        // 4. 用户选择一个放映计划
        const randomShowtime = showtimes[Math.floor(Math.random() * showtimes.length)];
        const showtimeId = randomShowtime['id'];

        randomSleep(2, 5);

        // 5. 用户查询座位表
        const seatMapRes = http.get(
            `${BASE_URL}/api/v1/showtimes/${showtimeId}/seatmap`,
            { headers: authHeaders }
        );
        checkResponse(seatMapRes, `Get showtime seatmap`, [200, 400]);

        if (seatMapRes.status !== 200) return;

        // 6. 用户有10% 的概率预定座位
        if (Math.random() < 1) {
            const seats = seatMapRes?.json()?.['seats'] ?? [];
            // 可用座位状态为0
            const availableSeats = seats.filter(seat => seat['status'] === 0);

            if (availableSeats.length > 0) {
                // 随机选择1-2个座位
                const seatsToBookCount = Math.floor(Math.random() * 2) + 1;
                const selectedSeats = availableSeats.slice(0, seatsToBookCount);
                const seatIdsToBook = selectedSeats.map(seat => seat['id']);

                if (seatIdsToBook.length > 0) {
                    const bookingPayload = JSON.stringify({
                        showtime_id: showtimeId,
                        seat_ids: seatIdsToBook,
                    });

                    const bookingRes = http.post(
                        `${BASE_URL}/api/v1/bookings`,
                        bookingPayload,
                        { headers: { ...authHeaders, 'Content-Type': 'application/json' } }
                    );

                    // 预定成功返回201，冲突可能返回400/409/500等，都视为符合预期的测试场景
                    checkResponse(bookingRes, `Book seats`, [201, 400]);
                }
            }
        }
        randomSleep(5, 10); // 完成操作或离开
    });
}
