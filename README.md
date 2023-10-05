# JsonGenius

## Description
JsonGenius allows you to extract data from any website using JSON Schema. It is a simple API that takes a JSON Schema and a URL and returns the data from the website in JSON format.

## Running
```bash
git clone https://github.com/semanser/jsongenius
cd jsongenius
docker compose up
```
The API will be available at http://localhost:3001.


## Example
#### Request
```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "url": "https://www.amazon.com/s?k=gaming+headsets",
  "schema": {
    "type": "object",
    "properties": {
      "products": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string",
              "description": "The product name"
            },
            "price": {
              "type": "number",
              "description": "The price of the product in USD"
            }
          }
        }
      }
    }
  }
}' http://localhost:3001/lookup
```

#### Response
```json
{
  "result": {
    "products": [
      {
        "name": "Razer Nari Ultimate Wireless 7.1 Surround Sound Gaming Headset: THX Audio & Haptic Feedback - Auto-Adjust Headband - Chroma RGB - Retractable Mic - For PC, PS4, PS5 - Black",
        "price": 99.52
      },
      {
        "name": "Razer BlackShark V2 Pro Wireless Gaming Headset 2023 Edition: Detachable Mic - Pro-Tuned FPS Profiles - 50mm Drivers - Noise-Isolating Earcups w/Ultra-Soft Memory Foam - 70 Hr Battery Life - Black",
        "price": 199
      },
      {
        "name": "BENGOO G9000 Stereo Gaming Headset for PS4 PC Xbox One PS5 Controller, Noise Cancelling Over Ear Headphones with Mic, LED Light, Bass Surround, Soft Memory Earmuffs for Laptop Mac Nintendo NES Games",
        "price":21.99
      },
      {
        "name": "SteelSeries Arctis Nova 1P Multi-System Gaming Headset — Hi-Fi Drivers — 360° Spatial Audio — Comfort Design — Durable — Lightweight — Noise-Cancelling Mic — PS5/PS4, PC, Xbox, Switch - White",
        "price":49.99
      },
      {
        "name": "HyperX Cloud II Gaming Headset - 7.1 Surround Sound - Memory Foam Ear Pads - Durable Aluminum Frame - Works with PC, PS4, PS4 PRO, Xbox One, Xbox One S - Gun Metal (KHX-HSCP-GM)",
        "price":77
      }
    ]
  }
}
```
