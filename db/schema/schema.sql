CREATE TABLE partido(
	id_partido		SERIAL,
	fecha			date		NOT NULL,
	tipo_cancha		int		NOT NULL,
	rival			varchar(50)	NOT NULL,
	resultado		varchar(20)	NOT NULL,
	posicion		varchar(10),	
	goles_propios		int		NOT NULL,
	asistencias_propias	int		NOT NULL,
	puntaje_propio		int		NOT NULL,
CONSTRAINT PK_PARTIDO PRIMARY KEY (id_partido)
);
