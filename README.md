# PharmPixel
Модуль интеграции с PharmaSpace

Для установки Pixel надо проделать следующие действия:


1. Скачать pixel https://s3-eu-west-1.amazonaws.com/updater.pharmecosystem.ru/pixel.zip
1. Распаковать в папку (пример C:\pixel)
1. Внести правки в файл install.bat
   - MP_USERNAME  -логин от личного кабинета pharmspace.ru
   - MP_PASSWORD - пароль от лично кабинета pharmspace.ru
   - OFD_TYPE - тип вашей системы OFD
   - OFD_TOKEN - логин и пароль
   - Если есть другие аккаунты ОФД тогда:
   - OFD_TYPE_1 - тип вашей системы OFD
   - OFD_TOKEN_1 -логин и пароль
   - Запустить командную строку cmd.exe из под Администратора
1. Перейти в папку с Pixel в пункте 2 и запустить install.bat
1. Если будут вопросы обращайтесь

 
Описание формата OFD_TOKEN для разных ОФД:

1. Type: ofdru Token: инн:логин:пароль  (пример 3245001416:mikhail.merkulov@megafon.ru:121212)
1. Type: ofd-ya Token: Для получения токена от ОФД вам надо зайти в личный кабинет на сайте ofd-ya.ru перейти в Профиль(верхний правый угол) перейти в раздел API и там найдете Ключ доступа который надо вставить в поле OFD_TOKEN
1. Type: 1ofd Token: логин:пароль
1. Type: taxcom Token: логин:пароль:ИД интегратора(можно полчить написав письмо в техподержку Такском)
1. Type: platformofd Token: логин:пароль
1. Type: sbis Token: инн:логин:пароль (пример 3245001416:mikhail.merkulov@megafon.ru:121212)
