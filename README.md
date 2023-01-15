# RabbitMQtoBitrix24


1 доработал программу, сделал реконект если соединение было разорвано
2 провел небольшое нагрузочное тестирование
2.1 создал на портале 10 000 слелок
2.2 отправил в очередь 10 000 сообщений на запрос методом crm.deal.list с выборкой по частичному совпадению названия сделки. это довольно сложная выборка из бд, хоть и возвращает только несколько элементов. все запросы залетели в очередь за 28 секунд
2.3 далее программа которая транслирует запросы в rest и публикует ответы справилась успешно, так же как и наш тестовый Б24 это успешно выдержал эти 10 000 запросов к rest api и выплюнул все результаты за 18 минут

Создание сервиса для автозапуска/перезапуска программы проброса запросов из RabbitMQ в REST Bitrix24.
<pre>
На сервере с правами root выполняем команды:
1 создаем файл /lib/systemd/system/b24rmq.service
nano /lib/systemd/system/b24rmq.service

2 помещаем в него следующее содержимое
[Unit]
Description=B24RMQ
After=network.target
[Service]
Environment="AMPQURL=amqp://login:pass@adr"
Environment="B24REST=https://.../rest/"
Environment="AMPQIN=bitrix24in"
Environment="AMPQOUT=bitrix24out"
Type=idle
User=root
ExecStart=/home/bitrix/B24RMQ
Restart=on-failure
RestartSec=10s
[Install]
WantedBy=multi-user.target


3 назначаем файлу права 644
sudo chmod 644 /lib/systemd/system/b24rmq.service

4 перезапускам список демонов
sudo systemctl daemon-reload

5 включаем автозапуск демона
sudo systemctl enable b24rmq.service

6 запускаем демона
sudo systemctl start b24rmq.service

</pre>



Изменения по формату обмена
<pre>
В целом он остался прежний. Тоесть документация по REST Bitrix24 актуальна, так же как и актуальны самописные методы REST
Обмен идет через очереди bitrix24in и bitrix24out
В очередь bitrix24in отправляем данные для REST запроса, из очереди bitrix24out читаем ответ по запросу. Спопставление запросов происходит по параметру message_id

В очередь bitrix24in передаем
Заголовки:
method – название метода REST. Например crm.lead.add (добавление лида)
user – пользователь из вебхука который мы вам предоставляли (https://.../rest/1/3w0fbh...lixzu1t/). в вашем случае 1
token – токен из вебхука который мы вам предоставляли (https://..../rest/1/3w0fbh3...zu1t/). в вашем случае 3w0fbh3itlixzu1t

Параметры:
message_id – уникальный идентификатор сообщения, он же возвращается и в ответе. по нему происходит сопоставление запроса с ответом
content_type – не обязателный. text/json
В тело сообщения передается json строка с параметрами REST запроса. Без изменений как и было.

В очередь bitrix24out попадает ответ REST

message_id – уникальный идентификатор сообщения, он же был передан в запросе. по нему происходит сопоставление запроса с ответом
в теле сообщения содержится json строка с ответом REST

</pre>
