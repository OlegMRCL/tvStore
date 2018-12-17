# tvStore

    GET /tv
Выводит список всех телевизоров из БД


Остальные запросы вовзращают упакованную в json структуру:

    type Answer struct {
      Ok     bool
      Err    error
      Tv     Tv
      Fields Validated
    }

Запросы:

    GET /tv/{id}
Возвращает по id из БД строку с конкретным телевизором

    POST /tv
    
    {"Brand" = "string", "Manufacturer" = "string", "Model" = "string", "Year" = 2011}
Создает в БД новую строку

    DELETE /tv/{id}
Удаляет из БД строку с указанным id

    PUT /tv/{id}
    {"Brand" = "string", "Manufacturer" = "string", "Model" = "string", "Year" = 2012}
Меняет значения полей строки в БД с указанным id на значения в теле запроса

Данные в запросах  "POST /tv" и "PUT /tv/{id}" проходят валидацию на соответсвие требованиям к конкретным полям:
  
  brand - может отсутствовать
  
  manufacturer - string, от 3 и более символов
  
  model - string, от 2 и более символов
  
  year - дата с 2010 года.
  

