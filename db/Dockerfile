# Imagen de la base de datos: PostgreSQL con el esquema incluido.
# Al crearse la base, Postgres ejecuta los .sql de /docker-entrypoint-initdb.d y crea las tablas.
# El archivo se copia con permiso de lectura para todos (--chmod=644), así funciona
# sin importar qué permisos tenga schema.sql en la máquina donde se ejecuta.
FROM postgres:alpine
COPY --chmod=644 schema/schema.sql /docker-entrypoint-initdb.d/
