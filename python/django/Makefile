.PHONY: start
start:
	python3 manage.py migrate cockroach_example && python3 manage.py runserver 6543

deps:
	git clone https://github.com/cockroachlabs/cockroach-django || true
	cd cockroach-django && pip3 install .
	pip3 install django
