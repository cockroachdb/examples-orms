.PHONY: start
start:
	python3 manage.py migrate cockroach_example && python3 manage.py runserver 6543

deps:
	pip3 install --upgrade setuptools
	pip3 install django==3.2.8 django-cockroachdb==3.2.1 psycopg2 dj-database-url==0.5.0
