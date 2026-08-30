import QRCode from "qrcode";

/** Render an otpauth URL as a PNG data URL for authenticator setup. */
export async function totpQrDataUrl(otpauthUrl: string): Promise<string> {
  return QRCode.toDataURL(otpauthUrl, {
    errorCorrectionLevel: "M",
    margin: 1,
    width: 192,
    color: { dark: "#111111", light: "#ffffff" },
  });
}
