# Proyecto: Shell de Windows para Controlar Linux Remotamente

## Descripción
Este proyecto consiste en una aplicación desarrollada en Go (Goland) que permite a un usuario en un equipo con Windows ejecutar comandos y controlar un sistema operativo Linux de forma remota. Utiliza protocolos de red seguros para establecer la comunicación entre ambos sistemas, permitiendo la automatización y administración de tareas de manera eficiente.

## Características
- Conexión remota desde Windows a Linux.
- Ejecución de comandos de terminal de Linux.
- Transferencia de archivos entre sistemas.
- Autenticación segura mediante SSH.
- Registro de actividades en logs.

## Requisitos del Sistema
- **Windows**: Sistema operativo Windows 10 o superior.
- **Linux**: Distribución basada en Unix con servidor SSH habilitado.
- **Herramientas y dependencias**:  
  - OpenSSH en el servidor Linux.
  - Go 1.18 o superior.

## Instalación
1. Clonar el repositorio:
   ```bash
   git clone https://github.com/usuario/windows-to-linux-shell.git
   cd windows-to-linux-shell
   ```
2. Compilar el proyecto:
   ```bash
   go build -o shell_remote
   ```

## Uso
1. Asegurarse de que el servidor Linux tiene SSH habilitado y accesible.
2. Ejecutar la aplicación en Windows:
   ```bash
   ./shell_remote
   ```
3. Ingresar la IP, usuario y contraseña del servidor Linux.
4. Ejecutar comandos directamente desde la interfaz o terminal.

## Ejemplo
```bash
> ls -la
> sudo systemctl restart apache2
```

## Seguridad
- Se recomienda usar claves SSH en lugar de contraseñas para mayor seguridad.
- Evitar almacenar información sensible en texto plano.

## Contribuciones
Las contribuciones son bienvenidas. Por favor, abre un issue antes de realizar un pull request para discutir cambios mayores.

