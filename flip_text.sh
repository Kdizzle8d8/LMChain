#!/bin/bash
flip_text() {
    local input="$1"
    local flipped=""
    local length=${#input}

    declare -A flip_map=(
        ["a"]="ɐ" ["b"]="q" ["c"]="ɔ" ["d"]="p" ["e"]="ǝ" ["f"]="ɟ" ["g"]="ƃ" ["h"]="ɥ"
        ["i"]="ᴉ" ["j"]="ɾ" ["k"]="ʞ" ["l"]="l" ["m"]="ɯ" ["n"]="u" ["o"]="o" ["p"]="d"
        ["q"]="b" ["r"]="ɹ" ["s"]="s" ["t"]="ʇ" ["u"]="n" ["v"]="ʌ" ["w"]="ʍ" ["x"]="x"
        ["y"]="ʎ" ["z"]="z" ["A"]="∀" ["B"]="𐐒" ["C"]="Ɔ" ["D"]="ᗡ" ["E"]="Ǝ" ["F"]="Ⅎ"
        ["G"]="⅁" ["H"]="H" ["I"]="I" ["J"]="ſ" ["K"]="⋊" ["L"]="˥" ["M"]="W" ["N"]="N"
        ["O"]="O" ["P"]="Ԁ" ["Q"]="Q" ["R"]="ᴚ" ["S"]="S" ["T"]="⊥" ["U"]="∩" ["V"]="Λ"
        ["W"]="M" ["X"]="X" ["Y"]="⅄" ["Z"]="Z" ["0"]="0" ["1"]="Ɩ" ["2"]="ᄅ" ["3"]="Ɛ"
        ["4"]="ㄣ" ["5"]="ϛ" ["6"]="9" ["7"]="ㄥ" ["8"]="8" ["9"]="6" ["!"]="¡" ["?"]="¿"
        ["."]="˙" [","]="'" ["'"]=","
    )

    for (( i=0; i<$length; i++ )); do
        char="${input:$i:1}"
        flipped="${flip_map[$char]:-$char}$flipped"
    done

    echo "$flipped"
}

if [ $# -eq 0 ]; then
    echo "Usage: $0 <text>"
    exit 1
fi

flip_text "$1"
