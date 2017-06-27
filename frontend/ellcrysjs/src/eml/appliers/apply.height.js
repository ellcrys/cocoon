import {AddToInlineStyle} from '../../util/util'
import cheerio from 'cheerio'

export type HeightApplier = {
    applyHeight: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMinHeight: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMaxHeight: ($el: cheerio.CheerioElement, defaultValue?: string) => void
}

export const Appliers: HeightApplier = {
   
   // applyHeight adds height
   applyHeight: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const height = $el.attribs["height"] || defaultValue
       delete $el.attribs["height"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `height:${height}`)
   },
   
   // applyMinHeight adds min-height
   applyMinHeight: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const minHeight = $el.attribs["minheight"] || defaultValue
       delete $el.attribs["minheight"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `min-height:${minHeight}`)
   },
   
   // applyMaxHeight adds max-height
   applyMaxHeight: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const maxHeight = $el.attribs["maxheight"] || defaultValue
       delete $el.attribs["maxheight"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `min-height:${maxHeight}`)
   },
}