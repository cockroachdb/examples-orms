.PHONY: start
start:
	python3 manage.py migrate cockroach_example && python3 manage.py runserver 6543

deps:
	git clone https://github.com/cockroachdb/django-cockroachdb || true
	pip3 install --upgrade setuptools
	pip3 install django
	cd django-cockroachdb && pip3 install .
