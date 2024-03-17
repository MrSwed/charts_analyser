# Charts analyser: web-service

Веб-сервис для мониторинга и анализа трэков судов и их соответствия
границам морских карт


## Запуск

Перед запуском необходимо подготовить импортируемые данные, см Импорт данных 

Запуск приложений с помощью docker composer
```sh
docker composer up
```
Первый запуск может длиться продолжительное время - в зависимости от количества импортируемых данных. Первым после доступности служебных сервисов (postgres), запускается миграция, сервер и симулятор ждут ее завершения

Адрес по умолчании для запускаемых приложений
 - Адрес сервера http://127.0.0.1:3000/api
 - Адрес swagger http://127.0.0.1:8000

## Импорт данных

По умолчанию, определены следующие папки
- `./data` - папка для файлов импорта
 Для импортирования данных зон и исторических треков, необходимо положить файл зон `geo_zones.json` в корень папки импорта. Формат файла: JSON массив с названиями зон в ключах и полигонами координат в значениях.
  При импорте предусмотрено исправление координат полигонов - если первая и последняя координаты не одинаковы - добавляется завершающая координата полигона
- `./data/tracks` - папка для набора треков - файлы *.csv с полями
  `timestamp,longitude,latitude,vessel_id,vessel_name`  


## Функциональность серверной части

#### Авторизация

Все роуты серверной части ограничены проверкой JWT токена.

В токене должна быть указана роль  

<details><summary>Click to expand</summary>

```json
{
  "sub":  "9233466",
  "name": "Saga Viking",
  "role": 1
}
```

</details>

- роль `1` - судно (отправка треков `POST /api/track`) 
- pоль `2` - оператор (управление судами, постановка/снятие на контроль, мониторинг)
- роль `4` - админ (управление операторами)

Получить токен для дальнейшей авторизации можно по роуту
- аутентификация `POST /api/login`, дальнейшая авторизация через токен

#### Роль Оператор
- список морских карт, которые пересекались заданными в запросе судами в заданный временной промежуток. `POST /api/chart/vessels`  
  Входные параметры: идентификаторы судов, стартовая дата, конечная дата. JSON в теле запроса
  <details><summary>Click to expand</summary>

  ```json
  {
   "vesselIDs": [
    8902967
   ],
   "start": "2017-10-12T00:00:11Z",
   "finish": "2017-10-13T00:00:11Z"
  }
  ```
  </details>  

- список судов, которые пересекали заданные в запросе морские карты в заданный временной промежуток. `POST /api/chart/vessels`  
  Входные параметры: идентификаторы карт, стартовая дата, конечная дата. JSON в теле запроса
  <details><summary>Click to expand</summary>

  ```json
  {
   "zoneNames": [
    "zone_205"
   ],
   "start": "2017-01-08T00:00:00Z",
   "finish": "2017-01-09T00:00:00Z"
  }
  ```
  </details>  
- Режим мониторинга судов в реальном времени:
  - поставить (снять) на мониторинг судно. `POST (DELETE) /api/monitor`
  - список судов, поставленных на мониторинг. `GET /api/monitor`
  - Детальная информация по судам на мониторинге: `POST /api/monitor/state`, кроме основной:
    - ID карты, которую судно пересекает в данный момент.
    - Время входа в текущую зону
    - Время, проведенное судном на мониторинге в текущей зоне (карте). Время
      отсчитывается с момента последнего пересечения границы зоны судном (момент
      входа в зону).
- добавление судов `POST /api/vessels`
- изменение  `PUT /api/vessels`
- удаление/восстановление  (soft delete) `DELETE/PATCH /api/vessels`
- GET `/api/track/:id` список треков за указанный период для судна

### Роль судно:

- идентификация судна, отправляющего трек, через токен
- POST `api/track` отправка трека судном. Запись в историю и отображение в мониторинге, если судно стоит на контроле. ID судна берется из JWT, исключая возможность ошибочной записи чужого трека

### Примечания:
- Морские карты (зоны) задаются полигонами с произвольным число вершин обозначенными географическими координатами.
- Считается, что судно пересекало карту, если хотя бы одна точка его маршрута
  находится в пределах полигона, описывающего карту.

## Ограничения
Сервис принимает треки только в следующих географических координатах:
- min_longitude = -180.0
- max_longitude = 180.0
- min_latitude = -75.0
- max_latitude = 75.0

## Симулятор

Симулятор потока реальных данных, который базируется на предоставленных
исторических данных.

Берет из истории несколько случайных судов, ставит их на контроль (`/api/monitor`) 
с токеном оператора, затем, с заданной периодичностью отправляет уже имеющиеся в базе треки с токеном судна (POST `/api/track`)

Может работать в режиме реального времени - отправляя трек в исторические часы, минуты.
По умолчанию отправляет треки раз в 10 сек.

