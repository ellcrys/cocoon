// @flow
import {Appliers as FlexboxAppliers} from './apply.flexbox'
import {Appliers as MarginAppliers} from './apply.margin'
import {Appliers as PaddingAppliers} from './apply.padding'
import {Appliers as HeightAppliers} from './apply.height'
import {Appliers as WidthAppliers} from './apply.width'

const appliers = {
    flexbox: FlexboxAppliers,
    margin: MarginAppliers,
    padding: PaddingAppliers,
    width: WidthAppliers,
    height: HeightAppliers
}

export default appliers