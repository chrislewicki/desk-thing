import time
import random
import adafruit_requests
import board
import displayio
import adafruit_touchscreen
import adafruit_hashlib as hashlib
import binascii
import xml.etree.ElementTree as ET
import busio
from digitalio import DigitalInOut
import adafruit_esp32spi.adafruit_esp32spi_socket as socket
import adafruit_esp32spi.adafruit_esp32spi as ESP

try:
    from secrets import secrets
except ImportError:
    print("wifi and other secrets are located in secrets.py, add them there!")
    raise

# ESP32 setup
esp32_cs = DigitalInOut(board.ESP_CS)
esp32_ready = DigitalInOut(board.ESP_BUSY)
esp32_reset = DigitalInOut(board.ESP_RESET)

spi = busio.SPI(board.SCK, board.MOSI, board.MISO)
esp = ESP.ESP_SPIcontrol(spi, esp32_cs, esp32_ready, esp32_reset)

print("Connecting to WiFi...")
if esp.status == ESP.WL_IDLE_STATUS:
    print("ESP32 found and in idle mode")
print("Firmware vers.", esp.firmware_version)
print("MAC addr:", [hex(i) for i in esp.MAC_address])

while not esp.is_connected:
    try:
        esp.connect_AP(secrets["ssid"], secrets["password"])
    except RuntimeError as e:
        print("could not connect to AP, retrying: ", e)
        continue
print("Connected to", str(esp.ssid, "utf-8"), "\tRSSI:", esp.rssi)

# Initialize requests with the ESP32 socket
requests = adafruit_requests.Session(socket, esp)

ts = adafruit_touchscreen.Touchscreen(
    board.TOUCH_XL, board.TOUCH_XR, board.TOUCH_YD, board.TOUCH_YU,
    calibration=((5200, 59000), (5800, 57000)), size=(320, 240)
)

def url_encode(s):
    """Simple URL encoding function"""
    safe_chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
    return "".join(c if c in safe_chars else "%" + hex(ord(c))[2:].upper().zfill(2) for c in str(s))

def sign(key, msg):
    """Simple HMAC-like signing function using hashlib"""
    block_size = 64  # Block size for SHA256
    if len(key) > block_size:
        key = hashlib.sha256(key).digest()
    if len(key) < block_size:
        key = key + b'\0' * (block_size - len(key))
    
    o_key_pad = bytes(x ^ 0x5c for x in key)
    i_key_pad = bytes(x ^ 0x36 for x in key)
    
    return hashlib.sha256(o_key_pad + hashlib.sha256(i_key_pad + msg.encode('utf-8')).digest()).digest()

def calculate_signature(secret_key, datestamp, region, string_to_sign, service='s3'):
    k_date = sign(('AWS4' + secret_key).encode('utf-8'), datestamp)
    k_region = sign(k_date, region)
    k_service = sign(k_region, service)
    k_signing = sign(k_service, 'aws4_request')
    signature = binascii.hexlify(sign(k_signing, string_to_sign)).decode('utf-8')
    return signature

def generate_canonical_request(method, canonical_uri, canonical_querystring, headers, payload_hash):
    canonical_headers = ''
    signed_headers = ''
    for header_key in sorted(headers.keys()):
        canonical_headers += header_key.lower() + ':' + headers[header_key].strip() + '\n'
        signed_headers += header_key.lower() + ';'
    signed_headers = signed_headers.rstrip(';')

    canonical_request = '\n'.join([
        method,
        canonical_uri,
        canonical_querystring,
        canonical_headers,
        signed_headers,
        payload_hash
    ])
    return canonical_request, signed_headers

def create_string_to_sign(canonical_request, amz_date, datestamp, region, service='s3'):
    algorithm = 'AWS4-HMAC-SHA256'
    credential_scope = f'{datestamp}/{region}/{service}/aws4_request'
    hashed_request = hashlib.sha256(canonical_request.encode('utf-8')).hexdigest()
    string_to_sign = '\n'.join([
        algorithm,
        amz_date,
        credential_scope,
        hashed_request
    ])
    return string_to_sign, credential_scope

def calculate_signature(secret_key, datestamp, region, string_to_sign, service='s3'):
    def sign(key, msg):
        return hmac.new(key, msg.encode('utf-8'), hashlib.sha256).digest()
    
    k_date = sign(('AWS4' + secret_key).encode('utf-8'), datestamp)
    k_region = sign(k_date, region)
    k_service = sign(k_region, service)
    k_signing = sign(k_service, 'aws4_request')
    signature = hmac.new(k_signing, string_to_sign.encode('utf-8'), hashlib.sha256).hexdigest()
    return signature

def get_authorization_header(access_key, credential_scope, signed_headers, signature):
    algorithm = 'AWS4-HMAC-SHA256'
    authorization_header = f'{algorithm} Credential={access_key}/{credential_scope}, SignedHeaders={signed_headers}, Signature={signature}'
    return authorization_header

def list_objects(prefix=''):
    access_key = secrets['do_space_access_key']
    secret_key = secrets['do_space_secret_key']
    region = secrets['do_space_region']
    space_name = secrets['do_space_name']
    host = f'{space_name}.{region}.digitaloceanspaces.com'

    method = 'GET'
    service = 's3'
    endpoint = f'https://{host}'
    canonical_uri = '/'
    payload_hash = hashlib.sha256(b'').hexdigest()

    query_params = {
        'list-type': '2'
    }
    if prefix:
        query_params['prefix'] = prefix
    
    canonical_querystring = '&'.join([f'{url_encode(k)}={url_encode(v)}' for k, v in sorted(query_params.items())])

    t = time.time()
    amz_date = time.strftime('%Y%m%dT%H%M%SZ', time.gmtime(t))
    datestamp = time.strftime('%Y%m%d', time.gmtime(t))

    headers = {
        'Host': host,
        'x-amz-content-sha256': payload_hash,
        'x-amz-date': amz_date,
    }

    canonical_request, signed_headers = generate_canonical_request(
        method,
        canonical_uri,
        canonical_querystring,
        headers,
        payload_hash
    )

    string_to_sign, credential_scope = create_string_to_sign(
        canonical_request,
        amz_date,
        datestamp,
        region,
        service
    )

    signature = calculate_signature(
        secret_key,
        datestamp,
        region,
        string_to_sign,
        service
    )

    authorization_header = get_authorization_header(
        access_key,
        credential_scope,
        signed_headers,
        signature
    )

    headers['Authorization'] = authorization_header

    # Requests library adds 'Host' to headers automatically
    headers.pop('Host', None)

    # Send GET request
    try:
        url = endpoint + '?' + canonical_querystring
        response = requests.request(
            method,
            url,
            headers=headers
        )
        if response.status_code == 200:
            # XML PARSING FUCK FUCK FUCK FUCK FUCK
            root = ET.fromstring(response.text)
            image_keys = []
            for contents in root.findall('.//Contents'):
                key = contents.find('Key').text
                image_keys.append(key)
            return image_keys
        else:
            print(f"Failed to list objects, status code: {response.status_code}")
            return []
    except Exception as e:
        print(f"Error listing objects: {e}")
        return []
    pass

def download_image(image_path):
    image_url = f"https://{secrets['do_space_name']}.{secrets['do_space_region']}.digitaloceanspaces.com/{image_path}"
    try:
        response = requests.get(image_url)
        if response.status_code == 200:
            # Save image
            with open("/sd/current_image.bmp", "wb") as f:
                f.write(response.content)
            return True
        else:
            print("failed to download image, status code:", response.status_code)
            return False
    except Exception as e:
        print("Error downloading image:", e)
        return False
    pass

def display_image(image_path):
    display = board.DISPLAY
    display.auto_refresh = False

    display_group = displayio.Group()
    display.show(display_group)

    try:
        # load the image
        with open(image_path, "rb") as f:
            odb = displayio.OnDiskBitmap(f)
            tile_grid = displayio.TileGrid(
                odb,
                pixel_shader=getattr(odb, 'pixel_shader', displayio.ColorConverter())
            )
            display_group.append(tile_grid)

        display.refresh()
        display.auto_refresh = True
    except Exception as e:
        print("Error displaying image:", e)
    pass

def delete_image(image_path):
    access_key = secrets['do_space_access_key']
    secret_key = secrets['do_space_secret_key']
    region = secrets['do_space_region']
    space_name = secrets['do_space_name']
    host = f'{space_name}.{region}.digitaloceanspaces.com'

    method = 'DELETE'
    service = 's3'
    endpoint = f'https://{host}/{image_path}'
    canonical_uri = '/' + url_encode(image_path)
    canonical_querystring = ''
    payload_hash = hashlib.sha256(b'').hexdigest()

    t = time.time()
    amz_date = time.strftime('%Y%m%dT%H%M%SZ', time.gmtime(t))
    datestamp = time.strftime('%Y%m%d', time.gmtime(t))

    # Headers
    headers = {
        'Host': host,
        'x-amz-content-sha256': payload_hash,
        'x-amz-date': amz_date,
    }

    # Generate Canonical Request
    canonical_request, signed_headers = generate_canonical_request(
        method,
        canonical_uri,
        canonical_querystring,
        headers,
        payload_hash
    )

    # Create String to Sign
    string_to_sign, credential_scope = create_string_to_sign(
        canonical_request,
        amz_date,
        datestamp,
        region,
        service
    )

    # Calculate Signature
    signature = calculate_signature(
        secret_key,
        datestamp,
        region,
        string_to_sign,
        service
    )

    # Create Authorization Header
    authorization_header = get_authorization_header(
        access_key,
        credential_scope,
        signed_headers,
        signature
    )

    headers['Authorization'] = authorization_header

    headers.pop('Host', None)

    # send DELETE request
    try:
        response = requests.request(
            method,
            endpoint,
            headers=headers
        )
        if response.status_code == 204:
            print(f"successfully deleted {image_path}")
            return True
        else:
            print(f"failed to delete {image_path}, status code: {response.status_code}")
            return False
    except Exception as e:
        print(f"Error deleting image: {e}")
        return False
    pass

def wait_for_acknowledgement(image_path):
    print("waiting for acknowledgement...")
    while True:
        touch_point = ts.touch_point
        if touch_point:
            print("screen touched at:", touch_point)
            if delete_image(image_path):
                print("image acknowledged and deleted")
            else:
                print("failed to delete image")
            break
        time.sleep(0.1)
    pass

def main():
    while True:
        important_images = list_objects(prefix='important/')
        if important_images:
            image_path = important_images[0]
            if download_image(image_path):
                display_image("/sd/current_image.bmp")
                wait_for_acknowledgement(image_path)
            else:
                time.sleep(60)
        else:
            # list regular images
            images = list_objects(prefix='images/')
            if images:
                # display a random image every 10 minutes
                image_path = random.choice(images)
                if download_image(image_path):
                    display_image("/sd/current_image.bmp")
                    time.sleep(600)
                else:
                    time.sleep(60)
            else:
                print("no images available")
                time.sleep(60)

if __name__ == "__main__":
    main()