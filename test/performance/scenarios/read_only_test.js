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
        read_only: {
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

// 生成放映计划查询参数
function generateShowtimeQueryParams(movieIds, cinemaHallIds) {
    const params = new URLSearchParams();

    // 40%概率添加电影ID
    if (Math.random() < 0.4 && movieIds && movieIds.length > 0) {
        const randomMovieId = movieIds[Math.floor(Math.random() * movieIds.length)];
        params.set('movie_id', randomMovieId.toString());
    }

    // 30%概率添加影院ID
    if (Math.random() < 0.3 && cinemaHallIds && cinemaHallIds.length > 0) {
        const randomCinemaHallId = cinemaHallIds[Math.floor(Math.random() * cinemaHallIds.length)];
        params.set('cinema_hall_id', randomCinemaHallId.toString());
    }

    // 30%概率添加日期
    if (Math.random() < 0.3) {
        // 生成今天到30天后的随机日期
        const today = new Date();
        const randomDays = Math.floor(Math.random() * 31); // 0-30天
        const randomDate = new Date(today);
        randomDate.setDate(today.getDate() + randomDays);
        // 格式化日期为 2006-01-02T15:04:05Z07:00 格式
        const formattedDate = randomDate.toISOString().replace(/\.\d+Z$/, 'Z');
        params.set('date', formattedDate);
    }

    // 添加分页参数
    params.set('page', '1');
    params.set('page_size', '20');
    return params.toString();
}

export default function () {
    group('Read Only API Tests', function () {
        const user = getRandomUser();
        const authHeaders = login(user['username'], user['password']);
        if (!authHeaders) return;

        // 模拟用户思考时间
        randomSleep(2, 6);

        // 获取电影列表
        const movieListRes = http.get(
            `${BASE_URL}/api/v1/movies?${generateMovieQueryParams()}`,
            { headers: authHeaders }
        );
        checkResponse(movieListRes, 'Get movies list');

        // 模拟用户思考时间
        randomSleep(2, 6);

        // 声明在外部以扩大作用域
        const movieIds = movieListRes?.json()?.['movies']?.map(movie => movie['id']) ?? [];
        if (movieListRes.status === 200) {
            // 随机选择一部电影查看详情
            if (movieIds.length > 0) {
                const randomMovieId = movieIds[Math.floor(Math.random() * movieIds.length)];
                const movieInfoRes = http.get(
                    `${BASE_URL}/api/v1/movies/${randomMovieId}`,
                    { headers: authHeaders }
                );
                checkResponse(movieInfoRes, `Get movie detail`);
            }
        }

        // 模拟用户思考时间
        randomSleep(2, 6);

        // 获取影院列表
        const cinemaListRes = http.get(
            `${BASE_URL}/api/v1/cinema-halls`,
            { headers: authHeaders }
        );
        checkResponse(cinemaListRes, 'Get cinemas list');

        // 模拟用户思考时间
        randomSleep(2, 6);

        // 声明在外部以扩大作用域
        const cinemaHallIds = cinemaListRes?.json()?.['cinema_halls']?.map(cinemaHall => cinemaHall['id']) ?? [];
        if (cinemaListRes.status === 200) {
            // 随机选择一个影院查看详情
            if (cinemaHallIds.length > 0) {
                const randomCinemaHallId = cinemaHallIds[Math.floor(Math.random() * cinemaHallIds.length)];
                const cinemaHallInfoRes = http.get(
                    `${BASE_URL}/api/v1/cinema-halls/${randomCinemaHallId}`,
                    { headers: authHeaders }
                );
                checkResponse(cinemaHallInfoRes, `Get cinema hall detail`);
            }
        }

        // 模拟用户思考时间
        randomSleep(2, 6);

        if (movieIds.length > 0 && cinemaHallIds.length > 0) {
            // 获取放映计划列表
            const showtimeListRes = http.get(
                `${BASE_URL}/api/v1/showtimes?${generateShowtimeQueryParams(movieIds, cinemaHallIds)}`,
                { headers: authHeaders }
            );
            checkResponse(showtimeListRes, 'Get showtimes list');

            // 模拟用户思考时间
            randomSleep(2, 6);

            // 随机选择一个放映计划查看详情
            if (showtimeListRes.status === 200) {
                const showtimeIds = showtimeListRes?.json()?.['showtimes']?.map(showtime => showtime['id']) ?? [];
                if (showtimeIds.length > 0) {
                    const randomShowtimeId = showtimeIds[Math.floor(Math.random() * showtimeIds.length)];

                    // 查询放映计划详情
                    const showtimeInfoRes = http.get(
                        `${BASE_URL}/api/v1/showtimes/${randomShowtimeId}`,
                        { headers: authHeaders }
                    );
                    checkResponse(showtimeInfoRes, `Get showtime detail`);

                    // 模拟用户思考时间
                    randomSleep(2, 6);

                    // 查询座位表
                    const seatMapRes = http.get(
                        `${BASE_URL}/api/v1/showtimes/${randomShowtimeId}/seatmap`,
                        { headers: authHeaders }
                    );
                    checkResponse(seatMapRes, `Get showtime seatmap`, [200, 400]);

                    // 模拟用户思考时间
                    randomSleep(2, 6);
                }
            }

        }
    })
}