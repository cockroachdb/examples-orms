.PHONY: start
start:
	./manage.py migrate cockroach_example && ./manage.py runserver 6543

deps:
	git clone https://github.com/cockroachlabs/cockroach-django || true
	cd cockroach-django && pip install .
	python -m pip install "django<2"
