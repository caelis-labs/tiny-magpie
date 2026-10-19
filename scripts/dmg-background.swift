// Packaging artwork generated with system typography and an SF Symbol.
import AppKit

let image = NSImage(size: NSSize(width: 660, height: 360))
image.lockFocus()
NSColor(calibratedWhite: 0.97, alpha: 1).setFill()
NSRect(x: 0, y: 0, width: 660, height: 360).fill()
func label(_ text: String, _ y: CGFloat, _ size: CGFloat, _ color: NSColor) {
    let p = NSMutableParagraphStyle()
    p.alignment = .center
    (text as NSString).draw(in: NSRect(x: 20, y: y, width: 620, height: 36),
        withAttributes: [.font: NSFont.systemFont(ofSize: size, weight: .medium),
                         .foregroundColor: color, .paragraphStyle: p])
}
if let arrow = NSImage(systemSymbolName: "arrow.right", accessibilityDescription: nil) {
    arrow.draw(in: NSRect(x: 303, y: 191, width: 54, height: 38))
}
label("拖到 Applications 即可安装 · Drag to install", 34, 13, .gray)
image.unlockFocus()
try image.tiffRepresentation!.write(to: URL(fileURLWithPath: CommandLine.arguments[1]))
