.PHONY: start
start:
	python3 manage.py migrate cockroach_example && python3 manage.py runserver 6543

deps:
	pip3 install --upgrade setuptools
	pip3 install django==3.1.4 django-cockroachdb==3.1.2 psycopg2
