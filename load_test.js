import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Конфигурация теста
export let options = {
    stages: [
        { duration: '1m', target: 100 },   // Разогрев до 100 RPS за 1 минуту
        { duration: '3m', target: 1000 },  // Поддержание 1000 RPS в течение 3 минут
        { duration: '1m', target: 0 },     // Плавное снижение нагрузки
    ],
    thresholds: {
        http_req_duration: ['p(99)<50'],   // 99% запросов должны выполняться быстрее 50 мс
        http_req_failed: ['rate<0.01'],    // Менее 1% запросов должно завершаться ошибками
    },
};

// Базовый URL API
const BASE_URL = 'http://localhost:8080';

// Каталог товаров
const MERCH_ITEMS = [
    "t-shirt", "cup", "book", "pen", "powerbank",
    "hoody", "umbrella", "socks", "wallet", "pink-hoody"
];

// Генерация пула пользователей
const USERS_COUNT = 10000;
const USERS = Array.from({ length: USERS_COUNT }, (_, i) => ({
    username: `user${i + 1}`,
    password: `password${i + 1}`,
}));

// Функция для аутентификации пользователя
function authenticate(user) {
    const payload = JSON.stringify({ username: user.username, password: user.password });
    const response = http.post(`${BASE_URL}/api/auth`, payload, {
        headers: { 'Content-Type': 'application/json' },
    });

    if (response.status !== 200) {
        throw new Error(`Authentication failed for ${user.username}`);
    }

    return response.json('token');
}

// Основной тестовый сценарий
export default function () {
    // Выбираем случайного пользователя из пула
    const user = USERS[randomIntBetween(0, USERS_COUNT - 1)];
    const token = authenticate(user);

    // Выбираем случайную операцию
    const operation = randomIntBetween(1, 3);
    switch (operation) {
        case 1:
            // Сценарий 1: Получение информации о пользователе
            const infoResponse = http.get(`${BASE_URL}/api/info`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            check(infoResponse, { 'Info status is 200': (r) => r.status === 200 });
            break;

        case 2:
            // Сценарий 2: Передача монет другому пользователю
            const toUser = USERS[randomIntBetween(0, USERS_COUNT - 1)].username;
            const sendCoinPayload = JSON.stringify({
                toUser,
                amount: randomIntBetween(1, 100), // Случайная сумма монет
            });
            const sendCoinResponse = http.post(`${BASE_URL}/api/sendCoin`, sendCoinPayload, {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });
            check(sendCoinResponse, { 'SendCoin status is 200': (r) => r.status === 200 });
            break;

        case 3:
            // Сценарий 3: Покупка товара
            const item = MERCH_ITEMS[randomIntBetween(0, MERCH_ITEMS.length - 1)];
            const buyItemResponse = http.get(`${BASE_URL}/api/buy/${item}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            check(buyItemResponse, { 'BuyItem status is 200': (r) => r.status === 200 });
            break;
    }

    // Пауза между запросами
    sleep(randomIntBetween(1, 3));
}