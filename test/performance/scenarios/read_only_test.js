import { group } from 'k6';
import http from 'k6/http';
import { 
    BASE_URL, 
    getRandomUser,
    login,
    checkResponse 
} from './utils.ts';

export const options = {
    scenarios: {
        read_only: {
            executor: 'constant-vus',
            vus: 20,           // 20个并发用户
            duration: '1s',      // 持续1秒
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95%的请求应该在500ms内完成
        errors: ['rate<0.1'],             // 错误率应该小于10%
    },
};

// 模拟电影类型
const genres = ['动作', '冒险', '喜剧', '剧情', '恐怖', '科幻', '动画', '纪录片'];

// 生成查询参数
function generateMovieQueryParams() {
    const params = new URLSearchParams();
    
    // 随机决定使用哪些参数
    const useGenre = Math.random() < 0.8;      // 80%概率使用类型搜索
    const useYear = Math.random() < 0.2;       // 20%概率使用年份搜索
    
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

function generateShowtimeQueryParams() {
    const params = new URLSearchParams();
    params.set('movie_id', '1');
    return params.toString();
}

export default function () {
    group('Read Only API Tests', function () {
        // 首先进行登录
        const user = getRandomUser();
        const authHeaders = login(user['username'], user['password']);
        if (!authHeaders) return;

        // 获取电影列表（带查询参数）
        const queryParams = generateMovieQueryParams();
        const movieListRes = http.get(
            `${BASE_URL}/api/v1/movies${queryParams}`,
            { headers: authHeaders }
        );
        checkResponse(movieListRes, 'Get movies list');
        
        if (movieListRes.status === 200) {
            const body = movieListRes.body ? JSON.parse(movieListRes.body) : null;
            const movies = body?.data;
            if (movies && movies.length > 0) {
                // 随机选择一部电影查看详情
                const randomMovie = movies[Math.floor(Math.random() * movies.length)];
                const movieDetailRes = http.get(
                    `${BASE_URL}/api/v1/movies/${randomMovie.id}`,
                    { headers: authHeaders }
                );
                checkResponse(movieDetailRes, 'Get movie detail');
            }
        }

        // 获取场次列表
        const showtimeListRes = http.get(`${BASE_URL}/api/v1/showtimes`, { headers: authHeaders });
        checkResponse(showtimeListRes, 'Get showtimes list');

        if (showtimeListRes.status === 200) {
            const body = showtimeListRes.body ? JSON.parse(showtimeListRes.body) : null;
            const showtimes = body?.data;
            if (showtimes && showtimes.length > 0) {
                // 随机选择一个场次查看座位图
                const randomShowtime = showtimes[Math.floor(Math.random() * showtimes.length)];
                const seatmapRes = http.get(
                    `${BASE_URL}/api/v1/showtimes/${randomShowtime.id}/seatmap`,
                    { headers: authHeaders }
                );
                checkResponse(seatmapRes, 'Get seat map');
            }
        }
    });
} 