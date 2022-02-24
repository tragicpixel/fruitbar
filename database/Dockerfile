FROM postgres AS dev
ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD fruitbar
ENV POSTGRES_DB fruitbar
COPY create_and_init.sql /docker-entrypoint-initdb.d/
# note that running this script will make postgres unavailable over tcp until it is completed, so dependent containers will have to wait

# port is off by one so you can access via pgadmin easily if you have a local postgres installation running
EXPOSE 5433 

# prod build ? (different credentials?)