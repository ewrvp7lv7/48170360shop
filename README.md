# Тестовое задание: Реализовать API для небольшого интернет магазина.

## Реализовано на базе:
*   Golang & PostgreSQL.
*   Основная логика запросов вынесена в функции (stored procedures) PostgreSQL.
*   Основой для API взят github.com/go-chi/chi. Выбран как самый быстрый по тестам.
*   Аутенификация токиеном на базе github.com/go-chi/jwtauth. Не лучший вариант, достаточно сырая.
*   Валидация при помощи github.com/go-ozzo/ozzo-validation реализована через методы (не через токиены), что снижает вероятность ошибки разработчика.
*   Провайдер БД github.com/jmoiron/sqlx.
*   В качестве ключа строк корзины взят тип UUID.
*   API протестировано Postmen.

## Возможности API:
Реализовано на связке REST+JSON.
### Для покупателя
*   Смотреть список товаров и цен.
*   Добавлять и убавлять товары в корзине. При вводе отрицательного значения количество товара уменьшается на эту величину. При нулевом количестве товар удаляется. При попытке получить отрицательное количество БД возвращает ошибку.
*   При этом покупателю возвращается список товаров в его корзине с промежуточной и общей суммой.
*   "Покупать товар" то есть уменьшать количество товара на складе на количество товара в корзине с одновременным ее удалением. \
### Для менеджера
*   Смотреть все корзины покупателей.
*   Добавлять в список товаров новый товар. При этом название товара является уникальным, нельзя добавить товар с тем же названием.
*   Изменять количество товаров на положительную или отрицательную величину. БД возвращает текущее значение, при попытке получить отрицательное количество возвращает ошибку.
*   Роли покупателя и менеджера ограничены, при попытке покупателя выполнить операции менеджера или наоборот БД вернет ошибку.





