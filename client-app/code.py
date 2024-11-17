import time
import board
import busio
import adafruit_requests as requests
from digitalio import DigitalInOut
from adafruit_esp32spi import adafruit_esp32spi

try:
    from secrets import secrets
except ImportError:
    print("Unable to import secrets. Do they exist?")
    raise

esp32_cs = DigitalInOut(board.ESP_CS)
esp32_ready = DigitalInOut(board.ESP_BUSY)
esp32_reset = DigitalInOut(board.ESP_RESET)

spi = busio.SPI(board.SCK, board.MOSI, board.MISO)

esp = adafruit_esp32spi.ESP_SPIcontrol(spi, esp32_cs, esp32_ready, esp32_reset)

print("connecting to wifi...")
while not esp.is_connected:
    try:
        esp.connect_AP(secrets['ssid'], secrets['password'])
    except RuntimeError as e:
        print("unable to connect. retrying... ", e)
        time.sleep(5)
print("connected")
print("internal ip is: ", esp.pretty_ip(esp.ip_address))

