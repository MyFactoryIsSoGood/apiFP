<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a name="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/MyFactoryIsSoGood/advisory_backend">
    <img src="https://i.imgur.com/RYkafWs.png" alt="Logo" width="350" height="80">
  </a>

<h3 align="center">Golang Fingerprint matching</h3>

  <p align="center">
    Веб-сервис для идентификации отпечатка пальца на GoLang
    <br>
  </p>
</div>

## О проекте

Сервис предоставляет эндпоинты, позволяющие проводить идентификацию и добавлять шаблоны отпечатка в базу. Отпечатки хранятся в формате ISO19794:2-2005
Детальное описание формата: https://templates.machinezoo.com/iso-19794-2-2005

Идентификация проводится по адресу /identify и требует form-data тела с ключом fingerprint и значением в виде .bmp изображения.
Подробнее можно посмотреть в doc.yml c помощью https://editor.swagger.io/

### Стек
При написании сервиса мы старались работать в рамках стандартной библиотеки. Исключением стала систем ORM gorm и пакет imaging, совместимый со стандартным image
* Go
* GORM
* imaging (https://github.com/disintegration/imaging)
* PostgreSQL
* Docker
* Docker Compose


<!-- GETTING STARTED -->
## Установка

Проект завернут в контейнер и вместе с базой данных оркестрируется через Docker Compose.
* Склонируйте репозиторий
* Выполните следующие команды:
```docker-compose build```, 
```docker-compose up```

Приложение будет доступно на 127.0.0.1:8080
