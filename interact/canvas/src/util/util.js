
export function AddToClassName(existing, newClass) {
    existing = existing || ''
    let existingList = existing.split(" ")
    existingList.push(newClass)
    return existingList.join(" ")
}

export function AddToInlineStyle(existing, newStyle) {
    existing = existing || ''
    let existingList = existing.split(";")
    existingList.push(newStyle)
    return existingList.join(";")
}