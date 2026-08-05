Write a function that acts as a simple character encoder. The function must take a single lowercase letter as input. If the letter is a vowel (a, e, i, o, u), replace it with its corresponding symbol from the mapping below. If it is a consonant, return the letter unchanged.

You must implement the following function:

func EncodeVowel(letter rune) rune {}

Mapping:
a ➔ @
e ➔ &
i ➔ !
o ➔ 0
u ➔ ^

Examples:
EncodeVowel('a') -> '@'
EncodeVowel('b') -> 'b'
EncodeVowel('e') -> '&'
